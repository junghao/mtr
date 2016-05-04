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

type dataLatency struct {
	sitePK   int
	dataType dataType
}

func (d *dataLatency) loadPK(r *http.Request) (res *result) {
	if d.dataType, res = loadDataType(r.URL.Query().Get("typeID")); !res.ok {
		return res
	}

	if d.sitePK, res = dataSitePK(r.URL.Query().Get("siteID")); !res.ok {
		return res
	}

	res = &statusOK

	return
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

	var t time.Time
	var mean, min, max, fifty, ninety int

	if mean, err = strconv.Atoi(r.URL.Query().Get("mean")); err != nil {
		return badRequest("invalid value for mean")
	}

	if r.URL.Query().Get("min") != "" {
		if min, err = strconv.Atoi(r.URL.Query().Get("min")); err != nil {
			return badRequest("invalid value for min")
		}
	}

	if r.URL.Query().Get("max") != "" {
		if max, err = strconv.Atoi(r.URL.Query().Get("max")); err != nil {
			return badRequest("invalid value for max")
		}
	}

	if r.URL.Query().Get("fifty") != "" {
		if fifty, err = strconv.Atoi(r.URL.Query().Get("fifty")); err != nil {
			return badRequest("invalid value for fifty")
		}
	}

	if r.URL.Query().Get("ninety") != "" {
		if ninety, err = strconv.Atoi(r.URL.Query().Get("ninety")); err != nil {
			return badRequest("invalid value for ninety")
		}
	}

	if t, err = time.Parse(time.RFC3339, r.URL.Query().Get("time")); err != nil {
		return badRequest("invalid time")
	}

	if res := d.loadPK(r); !res.ok {
		return res
	}

	if _, err = db.Exec(`INSERT INTO data.latency(sitePK, typePK, rate_limit, time, mean, min, max, fifty, ninety) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		d.sitePK, d.dataType.typePK, t.Truncate(time.Minute).Unix(), t, int32(mean), int32(min), int32(max), int32(fifty), int32(ninety)); err != nil {
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

	// TODO when add latency tags add delete here.

	if _, err = txn.Exec(`DELETE FROM data.latency_threshold WHERE sitePK = $1 AND typePK = $2`,
		d.sitePK, d.dataType.typePK); err != nil {
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

	var p ts.Plot

	resolution := r.URL.Query().Get("resolution")

	switch resolution {
	case "", "minute":
		resolution = "minute"
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-12), time.Now().UTC())
		p.SetXLabel("12 hours")
	case "five_minutes":
		resolution = "five_minutes"
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*3), time.Now().UTC())
		p.SetXLabel("48 hours")
	case "hour":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*28), time.Now().UTC())
		p.SetXLabel("4 weeks")
	default:
		return badRequest("invalid value for resolution")
	}

	if res := d.loadPlot(resolution, &p); !res.ok {
		return res
	}

	p.SetUnit(d.dataType.Unit)

	// TODO tags

	var lower, upper int
	var res *result

	if lower, upper, res = d.threshold(); !res.ok {
		return res
	}

	if !(lower == 0 && upper == 0) {
		p.SetThreshold(float64(lower)*d.dataType.Scale, float64(upper)*d.dataType.Scale)
	}
	//
	//var tags []string
	//
	//if tags, res = f.tags(); !res.ok {
	//	return res
	//}
	//
	//p.SetSubTitle("Tags: " + strings.Join(tags, ","))

	p.SetTitle(fmt.Sprintf("Site: %s - %s", r.URL.Query().Get("siteID"), strings.Title(d.dataType.Name)))

	var err error

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
loadPlot loads plot data.  Assumes f.loadPK has been called first.
*/
func (d *dataLatency) loadPlot(resolution string, p *ts.Plot) *result {
	var err error

	var rows *sql.Rows

	// TODO - loading avg(mean) at each resolution.  Need to add max(fifty) and max(ninety) when there are some values.

	switch resolution {
	case "minute":
		rows, err = dbR.Query(`SELECT date_trunc('`+resolution+`',time) as t, avg(mean) FROM data.latency WHERE
		sitePK = $1 AND typePK = $2
		AND time > now() - interval '12 hours'
		GROUP BY date_trunc('`+resolution+`',time)
		ORDER BY t ASC`,
			d.sitePK, d.dataType.typePK)
	case "five_minutes":
		rows, err = dbR.Query(`SELECT date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t,
		 avg(mean) FROM data.latency WHERE
		sitePK = $1 AND typePK = $2
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min'
		ORDER BY t ASC`,
			d.sitePK, d.dataType.typePK)
	case "hour":
		rows, err = dbR.Query(`SELECT date_trunc('`+resolution+`',time) as t, avg(mean) FROM data.latency WHERE
		sitePK = $1 AND typePK = $2
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time)
		ORDER BY t ASC`,
			d.sitePK, d.dataType.typePK)
	default:
		return internalServerError(fmt.Errorf("invalid resolution: %s", resolution))
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

	return &statusOK
}
