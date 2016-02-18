package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/mtr/ts"
	"github.com/lib/pq"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	late    = "magenta"
	bad     = "crimson"
	good    = "lawngreen"
	unknown = "lightskyblue"
)

var resolution = [...]string{
	"minute",
	"hour",
	"day",
}

var duration = [...]time.Duration{
	time.Minute,
	time.Hour,
	time.Hour * 24,
}

type fieldMetric struct {
	localityID, sourceID, typeID string
	localityName                 string
	localityPK, sourcePK, typePK int
	lower, upper                 int32
}

// TODO change to load.
func (f *fieldMetric) loadID(r *http.Request) *result {
	f.localityID = r.URL.Query().Get("localityID")
	f.sourceID = r.URL.Query().Get("sourceID")
	f.typeID = r.URL.Query().Get("typeID")

	if err := db.QueryRow(`SELECT localityPK,sourcePK, typePK, locality.name 
		FROM field.locality, field.type, field.source 
		WHERE localityID = $1 
		AND sourceID=$2
		AND typeID=$3`, f.localityID, f.sourceID, f.typeID).Scan(&f.localityPK, &f.sourcePK, &f.typePK, &f.localityName); err != nil {
		if err == sql.ErrNoRows {
			return &result{ok: false, code: http.StatusBadRequest, msg: "one or more of localityID, sourceID or typeID is invalid"}
		}
		return internalServerError(err)
	}

	return &statusOK
}

//  TODO - what to do about the latest values shown on the plot - read it from latest and add it seprately as the last value
//  TODO checkQuery on all the things.
//  TODO resolution parameter on plot and also CSV.
//  TODO lines on plots when min and max are different. Will have to drop color on points

func (f *fieldMetric) save(r *http.Request) *result {
	if res := checkQuery(r, []string{"localityID", "sourceID", "typeID", "time", "value"}, []string{}); !res.ok {
		return res
	}

	var err error

	var t time.Time
	var v int

	if v, err = strconv.Atoi(r.URL.Query().Get("value")); err != nil {
		return badRequest("invalid value")
	}

	if t, err = time.Parse(time.RFC3339, r.URL.Query().Get("time")); err != nil {
		return badRequest("invalid time")
	}

	if res := f.loadID(r); !res.ok {
		return res
	}

	var txn *sql.Tx

	if txn, err = db.Begin(); err != nil {
		return internalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM field.metric_latest 
		WHERE localityPK = $1
		AND sourcePK = $2
		AND typePK = $3`, f.localityPK, f.sourcePK, f.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if _, err = txn.Exec(`INSERT INTO field.metric_latest(localityPK, sourcePK, typePK, time, value) VALUES($1, $2, $3, $4, $5)`,
		f.localityPK, f.sourcePK, f.typePK, t.Truncate(time.Minute), int32(v)); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if err = txn.Commit(); err != nil {
		return internalServerError(err)
	}

	// insert and update the values in the minute, hour, and day tables
	for i, _ := range resolution {
		// Insert the value (which may already exist)
		if _, err = db.Exec(`INSERT INTO field.metric_`+resolution[i]+`(localityPK, sourcePK, typePK, time, min, max) VALUES($1, $2, $3, $4, $5, $6)`,
			f.localityPK, f.sourcePK, f.typePK, t.Truncate(duration[i]), int32(v), int32(v)); err != nil {
			if err, ok := err.(*pq.Error); ok && err.Code == `23505` {
				// ignore unique errors and then update.
			} else {
				return internalServerError(err)
			}
		}

		// update the min value
		if _, err = db.Exec(`UPDATE field.metric_`+resolution[i]+` SET min = $5
		WHERE localityPK = $1
		AND sourcePK = $2
		AND typePK = $3
		AND time = $4
		and min > $5`,
			f.localityPK, f.sourcePK, f.typePK, t.Truncate(duration[i]), int32(v)); err != nil {
			return internalServerError(err)

		}

		// update the max value
		if _, err = db.Exec(`UPDATE field.metric_`+resolution[i]+` SET max = $5
		WHERE localityPK = $1
		AND sourcePK = $2
		AND typePK = $3
		AND time = $4
		and max < $5`,
			f.localityPK, f.sourcePK, f.typePK, t.Truncate(duration[i]), int32(v)); err != nil {
			return internalServerError(err)
		}
	}

	return &statusOK
}

func (f *fieldMetric) delete(r *http.Request) *result {
	if res := checkQuery(r, []string{"localityID", "sourceID", "typeID"}, []string{}); !res.ok {
		return res
	}

	if res := f.loadID(r); !res.ok {
		return res
	}

	var err error
	var txn *sql.Tx

	if txn, err = db.Begin(); err != nil {
		return internalServerError(err)
	}

	for _, v := range resolution {
		if _, err = txn.Exec(`DELETE FROM field.metric_`+v+` WHERE localityPK = $1 AND sourcePK = $2 AND typePK = $3`,
			f.localityPK, f.sourcePK, f.typePK); err != nil {
			txn.Rollback()
			return internalServerError(err)
		}
	}

	if _, err = txn.Exec(`DELETE FROM field.metric_latest WHERE localityPK = $1 AND sourcePK = $2 AND typePK = $3`,
		f.localityPK, f.sourcePK, f.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if err = txn.Commit(); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func (f *fieldMetric) metricCSV(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"localityID", "sourceID", "typeID"}, []string{}); !res.ok {
		return res
	}

	if res := f.loadID(r); !res.ok {
		return res
	}

	var rows *sql.Rows
	var err error

	if rows, err = dbR.Query(`SELECT format('%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), min, max) as csv FROM field.metric_minute 
		WHERE localityPK = $1 AND sourcePK = $2 AND typePK = $3
		ORDER BY time ASC`,
		f.localityPK, f.sourcePK, f.typePK); err != nil && err != sql.ErrNoRows {
		return internalServerError(err)
	}
	defer rows.Close()

	var d string

	b.Write([]byte("date-time," + f.typeID))
	b.Write(eol)
	for rows.Next() {
		if err = rows.Scan(&d); err != nil {
			return internalServerError(err)
		}
		b.Write([]byte(d))
		b.Write(eol)
	}
	rows.Close()

	h.Set("Content-Disposition", `attachment; filename="MTR-`+strings.Replace(f.localityID+`-`+f.sourceID+`-`+f.typeID, " ", "-", -1)+`.csv"`)
	h.Set("Content-Type", "text/csv")

	return &statusOK
}

func (f *fieldMetric) svg(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"localityID", "sourceID", "typeID"}, []string{"plot"}); !res.ok {
		return res
	}

	if res := f.loadID(r); !res.ok {
		return res
	}

	var p ts.Plot

	if res := f.loadPlot(&p); !res.ok {
		return res
	}

	p.SetXAxis(time.Now().UTC().Add(time.Hour*-24), time.Now().UTC())

	var err error

	switch r.URL.Query().Get("plot") {
	case "spark":
		err = ts.SparkScatterLatest.Draw(p, b)
	case "":
		p.SetTitle(fmt.Sprintf("%s - %s - %s", f.localityName, f.sourceID, strings.Title(f.typeID)))

		switch f.typeID {
		case "voltage":
			p.SetUnit("V")
			p.SetYLabel("Voltage (V)")
			p.SetYAxis(0.0, 25.0)
		}

		err = ts.Scatter.Draw(p, b)
	}
	if err != nil {
		return internalServerError(err)
	}

	h.Set("Content-Type", "image/svg+xml")

	return &statusOK
}

/*
loadThreshold loads thresholds for the metric.  Assumes f.load has been called first.
*/
func (f *fieldMetric) loadThreshold() *result {
	//  there might be no threshold
	if err := dbR.QueryRow(`SELECT lower,upper FROM field.threshold
		WHERE localityPK = $1 AND sourcePK = $2 AND typePK = $3`,
		f.localityPK, f.sourcePK, f.typePK).Scan(&f.lower, &f.upper); err != nil && err != sql.ErrNoRows {
		return internalServerError(err)
	}

	return &statusOK
}

/*
loadPlot loads plot data.  Assumes f.load has been called first.
*/
func (f *fieldMetric) loadPlot(p *ts.Plot) *result {
	if res := f.loadThreshold(); !res.ok {
		return res
	}

	if !(f.lower == 0 && f.upper == 0) {
		switch f.typeID {
		case "voltage":
			p.SetThreshold(float64(f.lower)*0.001, float64(f.upper)*0.001)
		default:
			p.SetThreshold(float64(f.lower), float64(f.upper))
		}
	}

	var rows *sql.Rows
	var err error

	if rows, err = dbR.Query(`SELECT time, min,max FROM field.metric_minute
		WHERE localityPK = $1 AND sourcePK = $2 AND typePK = $3
		ORDER BY time ASC`,
		f.localityPK, f.sourcePK, f.typePK); err != nil {
		return internalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var min, max int32
	var pts []ts.Point

	for rows.Next() {
		if err = rows.Scan(&t, &min, &max); err != nil {
			return internalServerError(err)
		}
		pts = append(pts, ts.Point{DateTime: t, Value: float64(min), Colour: "darkcyan"})
		pts = append(pts, ts.Point{DateTime: t, Value: float64(max), Colour: "darkcyan"})
	}
	rows.Close()

	if f.typeID == "voltage" {
		for i, _ := range pts {
			pts[i].Value = pts[i].Value * 0.001
		}
	}

	p.AddSeries(ts.Series{Label: f.sourceID, Points: pts})

	return &statusOK
}
