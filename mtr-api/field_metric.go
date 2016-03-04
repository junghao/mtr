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
	deviceID, typeID string
	devicePK, typePK int
}

//  TODO what should this load and also caching?
func (f *fieldMetric) loadPK(r *http.Request) (res *result) {
	f.deviceID = r.URL.Query().Get("deviceID")
	f.typeID = r.URL.Query().Get("typeID")

	if f.devicePK, res = fieldDevicePK(f.deviceID); !res.ok {
		return
	}

	if f.typePK, res = fieldTypePK(f.typeID); !res.ok {
		return
	}

	res = &statusOK

	return
}

func (f *fieldMetric) save(r *http.Request) *result {
	if res := checkQuery(r, []string{"deviceID", "typeID", "time", "value"}, []string{}); !res.ok {
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

	if res := f.loadPK(r); !res.ok {
		return res
	}

	var txn *sql.Tx

	if txn, err = db.Begin(); err != nil {
		return internalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM field.metric_latest 
		WHERE devicePK = $1
		AND typePK = $2`, f.devicePK, f.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if _, err = txn.Exec(`INSERT INTO field.metric_latest(devicePK, typePK, time, value) VALUES($1, $2, $3, $4)`,
		f.devicePK, f.typePK, t.Truncate(time.Minute), int32(v)); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if err = txn.Commit(); err != nil {
		return internalServerError(err)
	}

	// insert and update the values in the minute, hour, and day tables
	for i, _ := range resolution {
		// Insert the value (which may already exist)
		if _, err = db.Exec(`INSERT INTO field.metric_`+resolution[i]+`(devicePK, typePK, time, min, max) VALUES($1, $2, $3, $4, $5)`,
			f.devicePK, f.typePK, t.Truncate(duration[i]), int32(v), int32(v)); err != nil {
			if err, ok := err.(*pq.Error); ok && err.Code == `23505` {
				// ignore unique errors and then update.
			} else {
				return internalServerError(err)
			}
		}

		// update the min value
		if _, err = db.Exec(`UPDATE field.metric_`+resolution[i]+` SET min = $4
		WHERE devicePK = $1
		AND typePK = $2
		AND time = $3
		and min > $4`,
			f.devicePK, f.typePK, t.Truncate(duration[i]), int32(v)); err != nil {
			return internalServerError(err)

		}

		// update the max value
		if _, err = db.Exec(`UPDATE field.metric_`+resolution[i]+` SET max = $4
		WHERE devicePK = $1
		AND typePK = $2
		AND time = $3
		and max < $4`,
			f.devicePK, f.typePK, t.Truncate(duration[i]), int32(v)); err != nil {
			return internalServerError(err)
		}
	}

	return &statusOK
}

func (f *fieldMetric) delete(r *http.Request) *result {
	if res := checkQuery(r, []string{"deviceID", "typeID"}, []string{}); !res.ok {
		return res
	}

	if res := f.loadPK(r); !res.ok {
		return res
	}

	var err error
	var txn *sql.Tx

	if txn, err = db.Begin(); err != nil {
		return internalServerError(err)
	}

	for _, v := range resolution {
		if _, err = txn.Exec(`DELETE FROM field.metric_`+v+` WHERE devicePK = $1 AND typePK = $2`,
			f.devicePK, f.typePK); err != nil {
			txn.Rollback()
			return internalServerError(err)
		}
	}

	if _, err = txn.Exec(`DELETE FROM field.metric_latest WHERE devicePK = $1 AND typePK = $2`,
		f.devicePK, f.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM field.metric_tag WHERE devicePK = $1 AND typePK = $2`,
		f.devicePK, f.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM field.threshold WHERE devicePK = $1 AND typePK = $2`,
		f.devicePK, f.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if err = txn.Commit(); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func (f *fieldMetric) metricCSV(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"deviceID", "typeID"}, []string{}); !res.ok {
		return res
	}

	if res := f.loadPK(r); !res.ok {
		return res
	}

	var rows *sql.Rows
	var err error

	if rows, err = dbR.Query(`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), min, max) as csv FROM field.metric_minute 
		WHERE devicePK = $1 AND typePK = $2
		ORDER BY time ASC`,
		f.devicePK, f.typePK); err != nil && err != sql.ErrNoRows {
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

	h.Set("Content-Disposition", `attachment; filename="MTR-`+strings.Replace(f.deviceID+`-`+f.typeID, " ", "-", -1)+`.csv"`)
	h.Set("Content-Type", "text/csv")

	return &statusOK
}

func (f *fieldMetric) svg(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"deviceID", "typeID"}, []string{"plot", "resolution", "yrange"}); !res.ok {
		return res
	}

	if res := f.loadPK(r); !res.ok {
		return res
	}

	var p ts.Plot

	resolution := r.URL.Query().Get("resolution")

	switch resolution {
	case "", "minute":
		resolution = "minute"
		p.SetXAxis(time.Now().UTC().Add(time.Minute*-1440), time.Now().UTC())
		p.SetXLabel("24 hours")
	case "hour":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*28), time.Now().UTC())
		p.SetXLabel("4 weeks")
	case "day":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*730), time.Now().UTC())
		p.SetXLabel("2 years")
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

		switch f.typeID {
		case "voltage":
			p.SetUnit("V")
			p.SetYLabel("Voltage (V)")
		}

		var lower, upper int
		var res *result

		if lower, upper, res = f.threshold(); !res.ok {
			return res
		}

		if !(lower == 0 && upper == 0) {
			switch f.typeID {
			case "voltage":
				p.SetThreshold(float64(lower)*0.001, float64(upper)*0.001)
			default:
				p.SetThreshold(float64(lower), float64(upper))
			}
		}

		var tags []string

		if tags, res = f.tags(); !res.ok {
			return res
		}

		p.SetTags(strings.Join(tags, ","))

		var mod string

		if mod, res = f.model(); !res.ok {
			return res
		}

		p.SetTitle(fmt.Sprintf("Device: %s, Model: %s, Metric: %s", f.deviceID, mod, strings.Title(f.typeID)))

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
func (f *fieldMetric) threshold() (lower, upper int, res *result) {
	res = &statusOK

	if err := dbR.QueryRow(`SELECT lower,upper FROM field.threshold
		WHERE devicePK = $1 AND typePK = $2`,
		f.devicePK, f.typePK).Scan(&lower, &upper); err != nil && err != sql.ErrNoRows {
		res = internalServerError(err)
	}

	return
}

func (f *fieldMetric) tags() (t []string, res *result) {
	var rows *sql.Rows
	var err error

	if rows, err = dbR.Query(`SELECT tag FROM field.metric_tag JOIN field.tag USING (tagpk) WHERE 
		devicePK = $1 AND typePK = $2
		ORDER BY tag asc`,
		f.devicePK, f.typePK); err != nil {
		res = internalServerError(err)
		return
	}

	defer rows.Close()

	var s string

	for rows.Next() {
		if err = rows.Scan(&s); err != nil {
			res = internalServerError(err)
			return
		}
		t = append(t, s)
	}

	res = &statusOK
	return
}

func (f *fieldMetric) model() (s string, res *result) {
	res = &statusOK

	if err := dbR.QueryRow(`SELECT modelid FROM field.device JOIN field.model using (modelpk)
		WHERE devicePK = $1`,
		f.devicePK).Scan(&s); err != nil && err != sql.ErrNoRows {
		res = internalServerError(err)
	}

	return
}

/*
loadPlot loads plot data.  Assumes f.load has been called first.
*/
func (f *fieldMetric) loadPlot(resolution string, p *ts.Plot) *result {
	var err error

	var latest ts.Point
	var latestValue int32

	if err = dbR.QueryRow(`SELECT time, value FROM field.metric_latest 
		WHERE devicePK = $1 AND typePK = $2`,
		f.devicePK, f.typePK).Scan(&latest.DateTime, &latestValue); err != nil {
		return internalServerError(err)
	}

	if f.typeID == "voltage" {
		latest.Value = float64(latestValue) * 0.001
	} else {
		latest.Value = float64(latestValue)
	}

	p.SetLatest(latest)

	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT time, min,max FROM field.metric_`+resolution+` WHERE 
		devicePK = $1 AND typePK = $2
		ORDER BY time ASC`,
		f.devicePK, f.typePK); err != nil {
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
			ptsMin = append(ptsMin, ts.Point{DateTime: t, Value: float64(min) * 0.001})
			ptsMax = append(ptsMax, ts.Point{DateTime: t, Value: float64(max) * 0.001})
		}
	} else {
		for rows.Next() {
			if err = rows.Scan(&t, &min, &max); err != nil {
				return internalServerError(err)
			}
			ptsMin = append(ptsMin, ts.Point{DateTime: t, Value: float64(min)})
			ptsMax = append(ptsMax, ts.Point{DateTime: t, Value: float64(max)})
		}
	}
	rows.Close()

	p.AddSeries(ts.Series{Label: f.deviceID, Points: ptsMin})
	p.AddSeries(ts.Series{Label: f.deviceID, Points: ptsMax})

	return &statusOK
}
