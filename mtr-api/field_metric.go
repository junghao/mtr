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
	deviceID  string
	devicePK  int
	fieldType fieldType
}

//  TODO what should this load and also caching?
func (f *fieldMetric) loadPK(r *http.Request) (res *result) {
	f.deviceID = r.URL.Query().Get("deviceID")

	if f.devicePK, res = fieldDevicePK(f.deviceID); !res.ok {
		return
	}

	if f.fieldType, res = loadFieldType(r.URL.Query().Get("typeID")); !res.ok {
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
		AND typePK = $2`, f.devicePK, f.fieldType.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if _, err = txn.Exec(`INSERT INTO field.metric_latest(devicePK, typePK, time, value) VALUES($1, $2, $3, $4)`,
		f.devicePK, f.fieldType.typePK, t.Truncate(time.Minute), int32(v)); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if err = txn.Commit(); err != nil {
		return internalServerError(err)
	}

	for i, _ := range resolution {
		// Insert the value (which may already exist)
		if _, err = db.Exec(`INSERT INTO field.metric_`+resolution[i]+`(devicePK, typePK, time, avg, n) VALUES($1, $2, $3, $4, $5)`,
			f.devicePK, f.fieldType.typePK, t.Truncate(duration[i]), int32(v), 1); err != nil {
			if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
				// unique error (already a value at this resolution) update the moving average.
				if _, err := db.Exec(`UPDATE field.metric_`+resolution[i]+` SET avg = ($4 + (avg * n)) / (n+1), n = n + 1
					WHERE devicePK = $1
					AND typePK = $2
					AND time = $3`,
					f.devicePK, f.fieldType.typePK, t.Truncate(duration[i]), int32(v)); err != nil {
					return internalServerError(err)
				}
			} else {
				return internalServerError(err)
			}
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
			f.devicePK, f.fieldType.typePK); err != nil {
			txn.Rollback()
			return internalServerError(err)
		}
	}

	if _, err = txn.Exec(`DELETE FROM field.metric_latest WHERE devicePK = $1 AND typePK = $2`,
		f.devicePK, f.fieldType.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM field.metric_tag WHERE devicePK = $1 AND typePK = $2`,
		f.devicePK, f.fieldType.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM field.threshold WHERE devicePK = $1 AND typePK = $2`,
		f.devicePK, f.fieldType.typePK); err != nil {
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

	if rows, err = dbR.Query(`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), avg, n) as csv FROM field.metric_minute 
		WHERE devicePK = $1 AND typePK = $2
		ORDER BY time ASC`,
		f.devicePK, f.fieldType.typePK); err != nil && err != sql.ErrNoRows {
		return internalServerError(err)
	}
	defer rows.Close()

	var d string

	b.Write([]byte("date-time," + f.fieldType.Name))
	b.Write(eol)
	for rows.Next() {
		if err = rows.Scan(&d); err != nil {
			return internalServerError(err)
		}
		b.Write([]byte(d))
		b.Write(eol)
	}
	rows.Close()

	h.Set("Content-Disposition", `attachment; filename="MTR-`+strings.Replace(f.deviceID+`-`+f.fieldType.Name, " ", "-", -1)+`.csv"`)
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
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-12), time.Now().UTC())
		p.SetXLabel("12 hours")
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

	p.SetUnit(f.fieldType.Unit)

	var lower, upper int
	var res *result

	if lower, upper, res = f.threshold(); !res.ok {
		return res
	}

	if !(lower == 0 && upper == 0) {
		p.SetThreshold(float64(lower)*f.fieldType.Scale, float64(upper)*f.fieldType.Scale)
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

	p.SetTitle(fmt.Sprintf("Device: %s, Model: %s, Metric: %s", f.deviceID, mod, strings.Title(f.fieldType.Name)))

	switch r.URL.Query().Get("plot") {
	case "spark", "spark-line":
		err = ts.SparkLine.Draw(p, b)
	case "spark-scatter":
		err = ts.SparkScatter.Draw(p, b)
	case "", "line":
		err = ts.Line.Draw(p, b)
	case "scatter":
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
func (f *fieldMetric) threshold() (lower, upper int, res *result) {
	res = &statusOK

	if err := dbR.QueryRow(`SELECT lower,upper FROM field.threshold
		WHERE devicePK = $1 AND typePK = $2`,
		f.devicePK, f.fieldType.typePK).Scan(&lower, &upper); err != nil && err != sql.ErrNoRows {
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
		f.devicePK, f.fieldType.typePK); err != nil {
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
		f.devicePK, f.fieldType.typePK).Scan(&latest.DateTime, &latestValue); err != nil {
		return internalServerError(err)
	}

	latest.Value = float64(latestValue) * f.fieldType.Scale

	p.SetLatest(latest)

	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT time, avg FROM field.metric_`+resolution+` WHERE 
		devicePK = $1 AND typePK = $2
		ORDER BY time ASC`,
		f.devicePK, f.fieldType.typePK); err != nil {
		return internalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var avg int32
	var pts []ts.Point

	for rows.Next() {
		if err = rows.Scan(&t, &avg); err != nil {
			return internalServerError(err)
		}
		pts = append(pts, ts.Point{DateTime: t, Value: float64(avg) * f.fieldType.Scale})
	}
	rows.Close()

	pts = append(pts, latest)

	p.AddSeries(ts.Series{Label: f.deviceID, Points: pts})

	return &statusOK
}
