package main

import (
	"net/http"
	"strconv"
	"time"
	"database/sql"
	"github.com/GeoNet/mtr/ts"
	"strings"
	"fmt"
	"bytes"
)

type dataLatency struct{
	sitePK int
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

	// Update or save the latest value.  Not rate limited.
	// TODO switch to Postgres 9.5 and use upsert.
	if _, err = db.Exec(`INSERT INTO data.latency_latest(sitePK, typePK, time, mean, min, max, fifty, ninety) VALUES($1, $2, $3, $4, $5, $6, $7, $8)`,
		d.sitePK, d.dataType.typePK, t, int32(mean), int32(min), int32(max), int32(fifty), int32(ninety)); err != nil {
		if _, err = db.Exec(`UPDATE data.latency_latest SET time = $3, mean = $4, min = $5, max = $6, fifty = $7, ninety = $8
				WHERE sitePK = $1
				AND typePK = $2
				AND time <= $3`,
			d.sitePK, d.dataType.typePK, t, int32(mean), int32(min), int32(max), int32(fifty), int32(ninety)); err != nil {
			return internalServerError(err)
		}
	}

	// Rate limit the stored data to 1 per minute
	var count int
	if err = db.QueryRow(`SELECT count(*) FROM data.latency
				WHERE sitePK = $1
				AND typePK = $2
				AND date_trunc('minute', time) = $3`, d.sitePK, d.dataType.typePK, t.Truncate(time.Minute)).Scan(&count); err != nil {
		if err != nil {
			return internalServerError(err)
		}
	}

	if count != 0 {
		return &statusTooManyRequests
	}

	if _, err = db.Exec(`INSERT INTO data.latency(sitePK, typePK, time, mean, min, max, fifty, ninety) VALUES($1, $2, $3, $4, $5, $6, $7, $8)`,
		d.sitePK, d.dataType.typePK, t, int32(mean), int32(min), int32(max), int32(fifty), int32(ninety)); err != nil {
		return internalServerError(err)
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

	if _, err = txn.Exec(`DELETE FROM data.latency_latest WHERE sitePK = $1 AND typePK = $2`,
		d.sitePK, d.dataType.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	// TODO when add latency thresholds look at this.
	// TODO also tags
	//if _, err = txn.Exec(`DELETE FROM data.threshold WHERE devicePK = $1 AND typePK = $2`,
	//	f.devicePK, f.fieldType.typePK); err != nil {
	//	txn.Rollback()
	//	return internalServerError(err)
	//}

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

	// TODO tags and thresholds

	//var lower, upper int
	//var res *result
	//
	//if lower, upper, res = f.threshold(); !res.ok {
	//	return res
	//}
	//
	//if !(lower == 0 && upper == 0) {
	//	p.SetThreshold(float64(lower)*f.fieldType.Scale, float64(upper)*f.fieldType.Scale)
	//}
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
	if err = dbR.QueryRow(`SELECT time, mean FROM data.latency_latest WHERE
			sitePK = $1 AND typePK = $2`,
		d.sitePK, d.dataType.typePK).Scan(&t, &value); err != nil {
		return internalServerError(err)
	}

	pts = append(pts, ts.Point{DateTime: t, Value: float64(value) * d.dataType.Scale})
	p.SetLatest(ts.Point{DateTime: t, Value: float64(value) * d.dataType.Scale}, "deepskyblue")

	p.AddSeries(ts.Series{Colour: "deepskyblue", Points: pts})

	return &statusOK
}

