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

func (a *dataLatency) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID", "time", "mean"}, []string{"min", "max", "fifty", "ninety"}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	var err error

	if a.mean, err = strconv.Atoi(v.Get("mean")); err != nil {
		return weft.BadRequest("invalid value for mean")
	}

	if v.Get("min") != "" {
		if a.min, err = strconv.Atoi(v.Get("min")); err != nil {
			return weft.BadRequest("invalid value for min")
		}
	}

	if v.Get("max") != "" {
		if a.max, err = strconv.Atoi(v.Get("max")); err != nil {
			return weft.BadRequest("invalid value for max")
		}
	}

	if v.Get("fifty") != "" {
		if a.fifty, err = strconv.Atoi(v.Get("fifty")); err != nil {
			return weft.BadRequest("invalid value for fifty")
		}
	}

	if v.Get("ninety") != "" {
		if a.ninety, err = strconv.Atoi(v.Get("ninety")); err != nil {
			return weft.BadRequest("invalid value for ninety")
		}
	}

	if a.t, err = time.Parse(time.RFC3339, v.Get("time")); err != nil {
		return weft.BadRequest("invalid time")
	}

	if res := a.dataSiteType.read(v.Get("siteID"), v.Get("typeID")); !res.Ok {
		return res
	}

	if _, err := db.Exec(`INSERT INTO data.latency(sitePK, typePK, rate_limit, time, mean, min, max, fifty, ninety) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		a.dataSite.pk, a.dataType.pk, a.t.Truncate(time.Minute).Unix(),
		a.t, int32(a.mean), int32(a.min), int32(a.max), int32(a.fifty), int32(a.ninety)); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			return &statusTooManyRequests
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func (a *dataLatency) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	if res := a.dataSiteType.read(v.Get("siteID"), v.Get("typeID")); !res.Ok {
		return res
	}

	var txn *sql.Tx
	var err error

	if txn, err = db.Begin(); err != nil {
		return weft.InternalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM data.latency WHERE sitePK = $1 AND typePK = $2`,
		a.dataSite.pk, a.dataType.pk); err != nil {
		txn.Rollback()
		return weft.InternalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM data.latency_threshold WHERE sitePK = $1 AND typePK = $2`,
		a.dataSite.pk, a.dataType.pk); err != nil {
		txn.Rollback()
		return weft.InternalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM data.latency_tag WHERE sitePK = $1 AND typePK = $2`,
		a.dataSite.pk, a.dataType.pk); err != nil {
		txn.Rollback()
		return weft.InternalServerError(err)
	}

	if err = txn.Commit(); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (a *dataLatency) svg(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{"plot", "resolution", "yrange"}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	if res := a.dataSiteType.read(v.Get("siteID"), v.Get("typeID")); !res.Ok {
		return res
	}

	switch r.URL.Query().Get("plot") {
	case "", "line":
		resolution := v.Get("resolution")
		if resolution == "" {
			resolution = "minute"
		}
		if res := a.plot(resolution, b); !res.Ok {
			return res
		}
	default:
		if res := a.spark(b); !res.Ok {
			return res
		}
	}

	h.Set("Content-Type", "image/svg+xml")

	return &weft.StatusOK
}

/*
plot draws an svg plot to b.  Assumes f.loadPK has been called first.
*/
func (a *dataLatency) plot(resolution string, b *bytes.Buffer) *weft.Result {
	var p ts.Plot

	p.SetUnit(a.dataType.Unit)

	var lower, upper int
	var res *weft.Result

	if err := dbR.QueryRow(`SELECT lower,upper FROM data.latency_threshold
		WHERE sitePK = $1 AND typePK = $2`,
		a.dataSite.pk, a.dataType.pk).Scan(&lower, &upper); err != nil && err != sql.ErrNoRows {
		res = weft.InternalServerError(err)
	}

	if !(lower == 0 && upper == 0) {
		p.SetThreshold(float64(lower)*a.dataType.Scale, float64(upper)*a.dataType.Scale)
	}

	var tags []string

	if tags, res = a.tags(); !res.Ok {
		return res
	}

	p.SetSubTitle("Tags: " + strings.Join(tags, ","))

	p.SetTitle(fmt.Sprintf("Site: %s - %s", a.dataSite.id, strings.Title(a.dataType.Name)))

	var err error
	var rows *sql.Rows

	// TODO - loading avg(mean) at each resolution.  Need to add max(fifty) and max(ninety) when there are some values.

	switch resolution {
	case "minute":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-12), time.Now().UTC())
		p.SetXLabel("12 hours")

		rows, err = dbR.Query(`SELECT date_trunc('`+resolution+`',time) as t, avg(mean) FROM data.latency WHERE
		sitePK = $1 AND typePK = $2
		AND time > now() - interval '12 hours'
		GROUP BY date_trunc('`+resolution+`',time)
		ORDER BY t ASC`,
			a.dataSite.pk, a.dataType.pk)
	case "five_minutes":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*2), time.Now().UTC())
		p.SetXLabel("48 hours")

		rows, err = dbR.Query(`SELECT date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t,
		 avg(mean) FROM data.latency WHERE
		sitePK = $1 AND typePK = $2
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min'
		ORDER BY t ASC`,
			a.dataSite.pk, a.dataType.pk)
	case "hour":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*28), time.Now().UTC())
		p.SetXLabel("4 weeks")

		rows, err = dbR.Query(`SELECT date_trunc('`+resolution+`',time) as t, avg(mean) FROM data.latency WHERE
		sitePK = $1 AND typePK = $2
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time)
		ORDER BY t ASC`,
			a.dataSite.pk, a.dataType.pk)
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
		pts = append(pts, ts.Point{DateTime: t, Value: avg * a.dataType.Scale})
	}
	rows.Close()

	// Add the latest value to the plot - this may be different to the average at minute or hour resolution.
	t = time.Time{}
	var value int32
	if err = dbR.QueryRow(`SELECT time, mean FROM data.latency WHERE
			sitePK = $1 AND typePK = $2
			ORDER BY time DESC
			LIMIT 1`,
		a.dataSite.pk, a.dataType.pk).Scan(&t, &value); err != nil {
		return weft.InternalServerError(err)
	}

	pts = append(pts, ts.Point{DateTime: t, Value: float64(value) * a.dataType.Scale})
	p.SetLatest(ts.Point{DateTime: t, Value: float64(value) * a.dataType.Scale}, "deepskyblue")

	p.AddSeries(ts.Series{Colour: "deepskyblue", Points: pts})

	if err = ts.Line.Draw(p, b); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

/*
spark draws an svg spark line to b.  Assumes f.loadPK has been called first.
*/
func (a *dataLatency) spark(b *bytes.Buffer) *weft.Result {
	var p ts.Plot

	p.SetXAxis(time.Now().UTC().Add(time.Hour*-12), time.Now().UTC())

	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t,
		 avg(mean) FROM data.latency WHERE
		sitePK = $1 AND typePK = $2
		AND time > now() - interval '12 hours'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min'
		ORDER BY t ASC`,
		a.dataSite.pk, a.dataType.pk); err != nil {
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
		pts = append(pts, ts.Point{DateTime: t, Value: avg * a.dataType.Scale})
	}
	rows.Close()

	p.AddSeries(ts.Series{Colour: "deepskyblue", Points: pts})

	if err = ts.SparkLine.Draw(p, b); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

// tags returns tags for a.  Assumes loadPK has been called.
func (a *dataLatency) tags() (t []string, res *weft.Result) {
	var rows *sql.Rows
	var err error

	if rows, err = dbR.Query(`SELECT tag FROM data.latency_tag JOIN mtr.tag USING (tagpk) WHERE
		sitePK = $1 AND typePK = $2
		ORDER BY tag asc`,
		a.dataSite.pk, a.dataType.pk); err != nil {
		res = weft.InternalServerError(err)
		return
	}

	defer rows.Close()

	var s string

	for rows.Next() {
		if err = rows.Scan(&s); err != nil {
			res = weft.InternalServerError(err)
			return
		}
		t = append(t, s)
	}

	res = &weft.StatusOK
	return
}
