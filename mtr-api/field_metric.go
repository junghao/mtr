package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/mtr/ts"
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var statusTooManyRequests = weft.Result{Ok: false, Code: http.StatusTooManyRequests, Msg: "Already data for the minute"}

func (f *fieldMetric) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "typeID", "time", "value"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	var err error

	if f.val, err = strconv.Atoi(v.Get("value")); err != nil {
		return weft.BadRequest("invalid value")
	}

	if f.t, err = time.Parse(time.RFC3339, v.Get("time")); err != nil {
		return weft.BadRequest("invalid time")
	}

	if res := f.read(v.Get("deviceID"), v.Get("typeID")); !res.Ok {
		return res
	}

	if _, err = db.Exec(`INSERT INTO field.metric(devicePK, typePK, rate_limit, time, value) VALUES($1, $2, $3, $4, $5)`,
		f.fieldDevice.pk, f.fieldType.pk, f.t.Truncate(time.Minute).Unix(), f.t, int32(f.val)); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			return &statusTooManyRequests
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func (f *fieldMetric) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "typeID"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	if res := f.read(v.Get("deviceID"), v.Get("typeID")); !res.Ok {
		return res
	}

	var err error
	var txn *sql.Tx

	if txn, err = db.Begin(); err != nil {
		return weft.InternalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM field.metric WHERE devicePK = $1 AND typePK = $2`,
		f.fieldDevice.pk, f.fieldType.pk); err != nil {
		txn.Rollback()
		return weft.InternalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM field.metric_tag WHERE devicePK = $1 AND typePK = $2`,
		f.fieldDevice.pk, f.fieldType.pk); err != nil {
		txn.Rollback()
		return weft.InternalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM field.threshold WHERE devicePK = $1 AND typePK = $2`,
		f.fieldDevice.pk, f.fieldType.pk); err != nil {
		txn.Rollback()
		return weft.InternalServerError(err)
	}

	if err = txn.Commit(); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (f *fieldMetric) svg(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "typeID"}, []string{"plot", "resolution"}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	if res := f.read(v.Get("deviceID"), v.Get("typeID")); !res.Ok {
		return res
	}

	switch r.URL.Query().Get("plot") {
	case "", "line":
		resolution := r.URL.Query().Get("resolution")
		if resolution == "" {
			resolution = "minute"
		}
		if res := f.plot(resolution, b); !res.Ok {
			return res
		}
	default:
		if res := f.spark(b); !res.Ok {
			return res
		}
	}

	h.Set("Content-Type", "image/svg+xml")

	return &weft.StatusOK
}

/*
plot draws an svg plot to b.  Assumes f.load has been called first.
Valid values for resolution are 'minute', 'five_minutes', 'hour'.
*/
func (f *fieldMetric) plot(resolution string, b *bytes.Buffer) *weft.Result {
	var p ts.Plot

	p.SetUnit(f.fieldType.Unit)

	var rows *sql.Rows
	var err error
	var lower, upper int

	if err := dbR.QueryRow(`SELECT lower,upper FROM field.threshold
		WHERE devicePK = $1 AND typePK = $2`,
		f.fieldDevice.pk, f.fieldType.pk).Scan(&lower, &upper); err != nil && err != sql.ErrNoRows {
		return weft.InternalServerError(err)
	}

	if !(lower == 0 && upper == 0) {
		p.SetThreshold(float64(lower)*f.fieldType.Scale, float64(upper)*f.fieldType.Scale)
	}

	var tags []string

	if rows, err = dbR.Query(`SELECT tag FROM field.metric_tag JOIN mtr.tag USING (tagpk) WHERE
		devicePK = $1 AND typePK = $2
		ORDER BY tag asc`,
		f.fieldDevice.pk, f.fieldType.pk); err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	var s string

	for rows.Next() {
		if err = rows.Scan(&s); err != nil {
			return weft.InternalServerError(err)
		}
		tags = append(tags, s)
	}

	rows.Close()

	p.SetSubTitle("Tags: " + strings.Join(tags, ","))

	var mod string

	if err = dbR.QueryRow(`SELECT modelid FROM field.device JOIN field.model using (modelpk)
		WHERE devicePK = $1`,
		f.fieldDevice.pk).Scan(&s); err != nil && err != sql.ErrNoRows {
		return weft.InternalServerError(err)
	}

	p.SetTitle(fmt.Sprintf("Device: %s, Model: %s, Metric: %s", f.fieldDevice.id, mod, strings.Title(f.fieldType.Name)))



	switch resolution {
	case "minute":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-12), time.Now().UTC())
		p.SetXLabel("12 hours")

		rows, err = dbR.Query(`SELECT date_trunc('`+resolution+`',time) as t, avg(value) FROM field.metric WHERE
		devicePK = $1 AND typePK = $2
		AND time > now() - interval '12 hours'
		GROUP BY date_trunc('`+resolution+`',time)
		ORDER BY t ASC`,
			f.fieldDevice.pk, f.fieldType.pk)
	case "five_minutes":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*2), time.Now().UTC())
		p.SetXLabel("48 hours")

		rows, err = dbR.Query(`SELECT date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t,
		 avg(value) FROM field.metric WHERE
		devicePK = $1 AND typePK = $2
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min'
		ORDER BY t ASC`,
			f.fieldDevice.pk, f.fieldType.pk)
	case "hour":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*28), time.Now().UTC())
		p.SetXLabel("4 weeks")

		rows, err = dbR.Query(`SELECT date_trunc('`+resolution+`',time) as t, avg(value) FROM field.metric WHERE
		devicePK = $1 AND typePK = $2
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time)
		ORDER BY t ASC`,
			f.fieldDevice.pk, f.fieldType.pk)
	default:
		return weft.BadRequest("invalid resolution")
	}
	if err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var avg float64
	var pts []ts.Point

	for rows.Next() {
		if err = rows.Scan(&t, &avg); err != nil {
			return weft.InternalServerError(err)
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
		f.fieldDevice.pk, f.fieldType.pk).Scan(&t, &value); err != nil {
		return weft.InternalServerError(err)
	}

	pts = append(pts, ts.Point{DateTime: t, Value: float64(value) * f.fieldType.Scale})
	p.SetLatest(ts.Point{DateTime: t, Value: float64(value) * f.fieldType.Scale}, "deepskyblue")

	p.AddSeries(ts.Series{Colour: "deepskyblue", Points: pts})

	if err = ts.Line.Draw(p, b); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

// spark draws an svg spark line to b.  Assumes loadPK has been called already.
func (f *fieldMetric) spark(b *bytes.Buffer) *weft.Result {
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
		f.fieldDevice.pk, f.fieldType.pk); err != nil {
		return weft.InternalServerError(err)
	}
	defer rows.Close()

	var t time.Time
	var avg float64
	var pts []ts.Point

	for rows.Next() {
		if err = rows.Scan(&t, &avg); err != nil {
			return weft.InternalServerError(err)
		}
		pts = append(pts, ts.Point{DateTime: t, Value: avg * f.fieldType.Scale})
	}
	rows.Close()

	p.AddSeries(ts.Series{Colour: "deepskyblue", Points: pts})

	if err = ts.SparkLine.Draw(p, b); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
