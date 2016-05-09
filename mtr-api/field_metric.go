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

var statusTooManyRequests = result{ok: false, code: http.StatusTooManyRequests, msg: "Already data for the minute"}

// TODO this should contain a device and type
type fieldMetric struct {
	deviceID  string
	devicePK  int
	fieldType fieldType
}

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

	if _, err = db.Exec(`INSERT INTO field.metric(devicePK, typePK, rate_limit, time, value) VALUES($1, $2, $3, $4, $5)`,
		f.devicePK, f.fieldType.typePK, t.Truncate(time.Minute).Unix(), t, int32(v)); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			return &statusTooManyRequests
		} else {
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

	if _, err = txn.Exec(`DELETE FROM field.metric WHERE devicePK = $1 AND typePK = $2`,
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

func (f *fieldMetric) svg(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"deviceID", "typeID"}, []string{"plot", "resolution"}); !res.ok {
		return res
	}

	if res := f.loadPK(r); !res.ok {
		return res
	}

	switch r.URL.Query().Get("plot") {
	case "", "line":
		resolution := r.URL.Query().Get("resolution")
		if resolution == "" {
			resolution = "minute"
		}
		if res := f.plot(resolution, b); !res.ok {
			return res
		}
	default:
		if res := f.spark(b); !res.ok {
			return res
		}
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

	if rows, err = dbR.Query(`SELECT tag FROM field.metric_tag JOIN mtr.tag USING (tagpk) WHERE
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
plot draws an svg plot to b.  Assumes f.load has been called first.
Valid values for resolution are 'minute', 'five_minutes', 'hour'.
*/
func (f *fieldMetric) plot(resolution string, b *bytes.Buffer) *result {
	var p ts.Plot

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

	p.SetSubTitle("Tags: " + strings.Join(tags, ","))

	var mod string

	if mod, res = f.model(); !res.ok {
		return res
	}

	p.SetTitle(fmt.Sprintf("Device: %s, Model: %s, Metric: %s", f.deviceID, mod, strings.Title(f.fieldType.Name)))

	var rows *sql.Rows
	var err error

	switch resolution {
	case "minute":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-12), time.Now().UTC())
		p.SetXLabel("12 hours")

		rows, err = dbR.Query(`SELECT date_trunc('`+resolution+`',time) as t, avg(value) FROM field.metric WHERE
		devicePK = $1 AND typePK = $2
		AND time > now() - interval '12 hours'
		GROUP BY date_trunc('`+resolution+`',time)
		ORDER BY t ASC`,
			f.devicePK, f.fieldType.typePK)
	case "five_minutes":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*2), time.Now().UTC())
		p.SetXLabel("48 hours")

		rows, err = dbR.Query(`SELECT date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t,
		 avg(value) FROM field.metric WHERE
		devicePK = $1 AND typePK = $2
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min'
		ORDER BY t ASC`,
			f.devicePK, f.fieldType.typePK)
	case "hour":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*28), time.Now().UTC())
		p.SetXLabel("4 weeks")

		rows, err = dbR.Query(`SELECT date_trunc('`+resolution+`',time) as t, avg(value) FROM field.metric WHERE
		devicePK = $1 AND typePK = $2
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time)
		ORDER BY t ASC`,
			f.devicePK, f.fieldType.typePK)
	default:
		return badRequest("invalid resolution")
	}
	if err != nil {
		return internalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var avg float64
	var pts []ts.Point

	for rows.Next() {
		if err = rows.Scan(&t, &avg); err != nil {
			return internalServerError(err)
		}
		pts = append(pts, ts.Point{DateTime: t, Value: avg * f.fieldType.Scale})
	}
	rows.Close()

	// Add the latest value to the plot - this may be different to the average at minute or hour resolution.
	t = time.Time{}
	var value int32
	if err = dbR.QueryRow(`SELECT time, value FROM field.metric WHERE
			devicePK = $1 AND typePK = $2
			ORDER BY time DESC
			LIMIT 1`,
		f.devicePK, f.fieldType.typePK).Scan(&t, &value); err != nil {
		return internalServerError(err)
	}

	pts = append(pts, ts.Point{DateTime: t, Value: float64(value) * f.fieldType.Scale})
	p.SetLatest(ts.Point{DateTime: t, Value: float64(value) * f.fieldType.Scale}, "deepskyblue")

	p.AddSeries(ts.Series{Colour: "deepskyblue", Points: pts})

	if err = ts.Line.Draw(p, b); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

// spark draws an svg spark line to b.  Assumes loadPK has been called already.
func (f *fieldMetric) spark(b *bytes.Buffer) *result {
	var p ts.Plot

	p.SetXAxis(time.Now().UTC().Add(time.Hour*-12), time.Now().UTC())

	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t,
		 avg(value) FROM field.metric WHERE
		devicePK = $1 AND typePK = $2
		AND time > now() - interval '12 hours'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min'
		ORDER BY t ASC`,
		f.devicePK, f.fieldType.typePK); err != nil {
		return internalServerError(err)
	}
	defer rows.Close()

	var t time.Time
	var avg float64
	var pts []ts.Point

	for rows.Next() {
		if err = rows.Scan(&t, &avg); err != nil {
			return internalServerError(err)
		}
		pts = append(pts, ts.Point{DateTime: t, Value: avg * f.fieldType.Scale})
	}
	rows.Close()

	p.AddSeries(ts.Series{Colour: "deepskyblue", Points: pts})

	if err = ts.SparkLine.Draw(p, b); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}
