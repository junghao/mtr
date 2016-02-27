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

	if rows, err = dbR.Query(`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), min, max) as csv FROM field.metric_minute 
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
	if res := checkQuery(r, []string{"localityID", "sourceID", "typeID"}, []string{"plot", "resolution", "yrange"}); !res.ok {
		return res
	}

	if res := f.loadID(r); !res.ok {
		return res
	}

	var p ts.Plot

	resolution := r.URL.Query().Get("resolution")

	switch resolution {
	case "":
		resolution = "minute"
		p.SetXAxis(time.Now().UTC().Add(time.Minute*-1440), time.Now().UTC())
	case "minute":
		p.SetXAxis(time.Now().UTC().Add(time.Minute*-1440), time.Now().UTC())
	case "hour":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-1440), time.Now().UTC())
	case "day":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*1440), time.Now().UTC())
	default:
		return badRequest("invalid value for resolution")
	}

	var err error

	if r.URL.Query().Get("yrange") != "" {
		y := strings.Split(r.URL.Query().Get("yrange"), `,`)

		var ymin, ymax float64

		if len(y) != 2 {
			return badRequest("invalid yrange query param.")
		}
		if ymin, err = strconv.ParseFloat(y[0], 64); err != nil {
			return badRequest("invalid yrange query param.")
		}
		if ymax, err = strconv.ParseFloat(y[1], 64); err != nil {
			return badRequest("invalid yrange query param.")
		}
		p.SetYAxis(ymin, ymax)
	}

	if res := f.loadPlot(resolution, &p); !res.ok {
		return res
	}

	switch r.URL.Query().Get("plot") {
	case "spark":
		switch f.typeID {
		case "voltage":
			p.SetUnit("V")
		}

		err = ts.SparkBarsLatest.DrawBars(p, b)
	case "":
		p.SetTitle(fmt.Sprintf("%s - %s - %s", f.localityName, f.sourceID, strings.Title(f.typeID)))

		switch f.typeID {
		case "voltage":
			p.SetUnit("V")
			p.SetYLabel("Voltage (V)")
		}

		err = ts.Bars.DrawBars(p, b)
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
func (f *fieldMetric) loadPlot(resolution string, p *ts.Plot) *result {
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

	var err error

	var latest ts.Point
	var latestValue int32

	if err = dbR.QueryRow(`SELECT time, value FROM field.metric_latest 
		WHERE localityPK = $1 AND sourcePK = $2 AND typePK = $3`,
		f.localityPK, f.sourcePK, f.typePK).Scan(&latest.DateTime, &latestValue); err != nil {
		return internalServerError(err)
	}

	switch {
	case latest.DateTime.Before(time.Now().UTC().Add(time.Hour * -48)):
		latest.Colour = late
	case f.lower == 0 && f.upper == 0:
		latest.Colour = unknown
	case latestValue <= f.upper && latestValue >= f.lower:
		latest.Colour = good
	default:
		latest.Colour = bad
	}

	if f.typeID == "voltage" {
		latest.Value = float64(latestValue) * 0.001
	} else {
		latest.Value = float64(latestValue)
	}

	p.SetLatest(latest)

	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT time, min,max FROM field.metric_`+resolution+` WHERE 
		localityPK = $1 AND sourcePK = $2 AND typePK = $3
		ORDER BY time ASC`,
		f.localityPK, f.sourcePK, f.typePK); err != nil {
		return internalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var min, max int32
	var ptsMin []ts.Point
	var ptsMax []ts.Point

	if f.typeID == "voltage" {
		for rows.Next() {
			if err = rows.Scan(&t, &min, &max); err != nil {
				return internalServerError(err)
			}
			ptsMin = append(ptsMin, ts.Point{DateTime: t, Value: float64(min) * 0.001, Colour: "darkcyan"})
			ptsMax = append(ptsMax, ts.Point{DateTime: t, Value: float64(max) * 0.001, Colour: "darkcyan"})
		}
	} else {
		for rows.Next() {
			if err = rows.Scan(&t, &min, &max); err != nil {
				return internalServerError(err)
			}
			ptsMin = append(ptsMin, ts.Point{DateTime: t, Value: float64(min), Colour: "darkcyan"})
			ptsMax = append(ptsMax, ts.Point{DateTime: t, Value: float64(max), Colour: "darkcyan"})
		}
	}
	rows.Close()

	p.AddSeries(ts.Series{Label: f.sourceID, Points: ptsMin})
	p.AddSeries(ts.Series{Label: f.sourceID, Points: ptsMax})

	return &statusOK
}
