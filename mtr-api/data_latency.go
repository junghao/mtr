package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/mtr/ts"
	"github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type dataLatency struct {
	dataSite
	dataType
	pk                            *result // for tracking pkLoad()
	t                             time.Time
	mean, min, max, fifty, ninety int
}

// TODO adopt this no-op approach further?
// Also pass the http.Resquest every where?
func (d *dataLatency) loadPK(r *http.Request) *result {
	if d.pk == nil {
		if d.pk = d.dataType.load(r); !d.pk.ok {
			return d.pk
		}

		if d.pk = d.dataSite.loadPK(r); !d.pk.ok {
			return d.pk
		}

		d.pk = &statusOK
	}

	return d.pk
}

/*
loadThreshold loads thresholds for the data latency.  Assumes d.loadPK has been called first.
*/
func (d *dataLatency) threshold() (lower, upper int, res *result) {
	res = &statusOK

	if err := dbR.QueryRow(`SELECT lower,upper FROM data.latency_threshold
		WHERE sitePK = $1 AND typePK = $2`,
		d.sitePK, d.dataType.typePK).Scan(&lower, &upper); err != nil && err != sql.ErrNoRows {
		res = internalServerError(err)
	}

	return
}

func (d *dataLatency) save(r *http.Request) *result {
	if res := checkQuery(r, []string{"siteID", "typeID", "time", "mean"}, []string{"min", "max", "fifty", "ninety"}); !res.ok {
		return res
	}

	var err error

	if d.mean, err = strconv.Atoi(r.URL.Query().Get("mean")); err != nil {
		return badRequest("invalid value for mean")
	}

	if r.URL.Query().Get("min") != "" {
		if d.min, err = strconv.Atoi(r.URL.Query().Get("min")); err != nil {
			return badRequest("invalid value for min")
		}
	}

	if r.URL.Query().Get("max") != "" {
		if d.max, err = strconv.Atoi(r.URL.Query().Get("max")); err != nil {
			return badRequest("invalid value for max")
		}
	}

	if r.URL.Query().Get("fifty") != "" {
		if d.fifty, err = strconv.Atoi(r.URL.Query().Get("fifty")); err != nil {
			return badRequest("invalid value for fifty")
		}
	}

	if r.URL.Query().Get("ninety") != "" {
		if d.ninety, err = strconv.Atoi(r.URL.Query().Get("ninety")); err != nil {
			return badRequest("invalid value for ninety")
		}
	}

	if d.t, err = time.Parse(time.RFC3339, r.URL.Query().Get("time")); err != nil {
		return badRequest("invalid time")
	}

	if res := d.loadPK(r); !res.ok {
		return res
	}

	if _, err = db.Exec(`INSERT INTO data.latency(sitePK, typePK, rate_limit, time, mean, min, max, fifty, ninety) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		d.sitePK, d.dataType.typePK, d.t.Truncate(time.Minute).Unix(),
		d.t, int32(d.mean), int32(d.min), int32(d.max), int32(d.fifty), int32(d.ninety)); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			return &statusTooManyRequests
		} else {
			return internalServerError(err)
		}
	}

	return &statusOK
}

func (d *dataLatency) delete(r *http.Request) *result {
	if res := checkQuery(r, []string{"siteID", "typeID"}, []string{}); !res.ok {
		return res
	}

	var err error
	if res := d.loadPK(r); !res.ok {
		return res
	}

	var txn *sql.Tx

	if txn, err = db.Begin(); err != nil {
		return internalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM data.latency WHERE sitePK = $1 AND typePK = $2`,
		d.sitePK, d.dataType.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM data.latency_threshold WHERE sitePK = $1 AND typePK = $2`,
		d.sitePK, d.dataType.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM data.latency_tag WHERE sitePK = $1 AND typePK = $2`,
		d.sitePK, d.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if err = txn.Commit(); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func (d *dataLatency) svg(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"siteID", "typeID"}, []string{"plot", "resolution", "yrange"}); !res.ok {
		return res
	}

	if res := d.loadPK(r); !res.ok {
		return res
	}

	d.siteID = r.URL.Query().Get("siteID")

	switch r.URL.Query().Get("plot") {
	case "", "line":
		resolution := r.URL.Query().Get("resolution")
		if resolution == "" {
			resolution = "minute"
		}
		if res := d.plot(resolution, b); !res.ok {
			return res
		}
	default:
		if res := d.spark(b); !res.ok {
			return res
		}
	}

	h.Set("Content-Type", "image/svg+xml")

	return &statusOK
}

/*
plot draws an svg plot to b.  Assumes f.loadPK has been called first.
*/
func (d *dataLatency) plot(resolution string, b *bytes.Buffer) *result {
	var p ts.Plot

	p.SetUnit(d.dataType.Unit)

	var lower, upper int
	var res *result

	if lower, upper, res = d.threshold(); !res.ok {
		return res
	}

	if !(lower == 0 && upper == 0) {
		p.SetThreshold(float64(lower)*d.dataType.Scale, float64(upper)*d.dataType.Scale)
	}

	var tags []string

	if tags, res = d.tags(); !res.ok {
		return res
	}

	p.SetSubTitle("Tags: " + strings.Join(tags, ","))

	p.SetTitle(fmt.Sprintf("Site: %s - %s", d.siteID, strings.Title(d.dataType.Name)))

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
			d.sitePK, d.dataType.typePK)
	case "five_minutes":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*2), time.Now().UTC())
		p.SetXLabel("48 hours")

		rows, err = dbR.Query(`SELECT date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t,
		 avg(mean) FROM data.latency WHERE
		sitePK = $1 AND typePK = $2
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min'
		ORDER BY t ASC`,
			d.sitePK, d.dataType.typePK)
	case "hour":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*28), time.Now().UTC())
		p.SetXLabel("4 weeks")

		rows, err = dbR.Query(`SELECT date_trunc('`+resolution+`',time) as t, avg(mean) FROM data.latency WHERE
		sitePK = $1 AND typePK = $2
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time)
		ORDER BY t ASC`,
			d.sitePK, d.dataType.typePK)
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
		log.Print(t, avg)
		pts = append(pts, ts.Point{DateTime: t, Value: avg * d.dataType.Scale})
	}
	rows.Close()

	// Add the latest value to the plot - this may be different to the average at minute or hour resolution.
	t = time.Time{}
	var value int32
	if err = dbR.QueryRow(`SELECT time, mean FROM data.latency WHERE
			sitePK = $1 AND typePK = $2
			ORDER BY time DESC
			LIMIT 1`,
		d.sitePK, d.dataType.typePK).Scan(&t, &value); err != nil {
		return internalServerError(err)
	}

	pts = append(pts, ts.Point{DateTime: t, Value: float64(value) * d.dataType.Scale})
	p.SetLatest(ts.Point{DateTime: t, Value: float64(value) * d.dataType.Scale}, "deepskyblue")

	p.AddSeries(ts.Series{Colour: "deepskyblue", Points: pts})

	if err = ts.Line.Draw(p, b); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

/*
spark draws an svg spark line to b.  Assumes f.loadPK has been called first.
*/
func (d *dataLatency) spark(b *bytes.Buffer) *result {
	var p ts.Plot

	p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*3), time.Now().UTC())

	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t,
		 avg(mean) FROM data.latency WHERE
		sitePK = $1 AND typePK = $2
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min'
		ORDER BY t ASC`,
		d.sitePK, d.dataType.typePK); err != nil {
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
		pts = append(pts, ts.Point{DateTime: t, Value: avg * d.dataType.Scale})
	}
	rows.Close()

	p.AddSeries(ts.Series{Colour: "deepskyblue", Points: pts})

	if err = ts.SparkLine.Draw(p, b); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

// tags returns tags for f.  Assumes loadPK has been called.
func (f *dataLatency) tags() (t []string, res *result) {
	var rows *sql.Rows
	var err error

	if rows, err = dbR.Query(`SELECT tag FROM data.latency_tag JOIN mtr.tag USING (tagpk) WHERE
		sitePK = $1 AND typePK = $2
		ORDER BY tag asc`,
		f.sitePK, f.typePK); err != nil {
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
