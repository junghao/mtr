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

func dataCompletenessPut(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	v := r.URL.Query()

	var err error

	var t time.Time
	var count int

	if count, err = strconv.Atoi(v.Get("count")); err != nil {
		return weft.BadRequest("invalid value for count")
	}

	if t, err = time.Parse(time.RFC3339, v.Get("time")); err != nil {
		return weft.BadRequest("invalid time")
	}

	siteID := v.Get("siteID")
	typeID := v.Get("typeID")

	var result sql.Result

	if result, err = db.Exec(`INSERT INTO data.completeness(sitePK, typePK, rate_limit, time, count)
				SELECT sitePK, typePK, $3, $4, $5
				FROM data.site, data.completeness_type
				WHERE siteID = $1
				AND typeID = $2`,
		siteID, typeID, t.Truncate(time.Minute).Unix(), t, int32(count)); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			return &statusTooManyRequests
		} else {
			return weft.InternalServerError(err)
		}
	}

	var i int64
	if i, err = result.RowsAffected(); err != nil {
		return weft.InternalServerError(err)
	}
	if i != 1 {
		return weft.BadRequest("Didn't create row, check your query parameters exist")
	}

	// Update the summary values if the incoming is newer.
	if result, err = db.Exec(`UPDATE data.completeness_summary SET
				time = $3, count = $4
				WHERE time < $3
				AND sitePK = (SELECT sitePK from data.site WHERE siteID = $1)
				AND typePK = (SELECT typePK from data.completeness_type WHERE typeID = $2)`,
		siteID, typeID, t, int32(count)); err != nil {
		return weft.InternalServerError(err)
	}

	// If no rows change either the values are old or it's the first time we've seen this metric.
	if i, err = result.RowsAffected(); err != nil {
		return weft.InternalServerError(err)
	}
	if i != 1 {
		if _, err = db.Exec(`INSERT INTO data.completeness_summary(sitePK, typePK, time, count)
				SELECT sitePK, typePK, $3, $4
				FROM data.site, data.completeness_type
				WHERE siteID = $1
				AND typeID = $2`,
			siteID, typeID, t, int32(count)); err != nil {
			if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
				// incoming value was old
			} else {
				return weft.InternalServerError(err)
			}
		}
	}

	return &weft.StatusOK
}

func dataCompletenessDelete(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	v := r.URL.Query()

	siteID := v.Get("siteID")
	typeID := v.Get("typeID")

	var txn *sql.Tx
	var err error

	if txn, err = db.Begin(); err != nil {
		return weft.InternalServerError(err)
	}

	for _, table := range []string{"data.completeness", "data.completeness_summary", "data.completeness_tag"} {
		if _, err = txn.Exec(`DELETE FROM `+table+` WHERE
				sitePK = (SELECT sitePK FROM data.site WHERE siteID = $1)
				AND typePK = (SELECT typePK FROM data.completeness_type WHERE typeID = $2)`,
			siteID, typeID); err != nil {
			txn.Rollback()
			return weft.InternalServerError(err)
		}
	}

	if err = txn.Commit(); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func dataCompletenessSvg(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	v := r.URL.Query()

	switch r.URL.Query().Get("plot") {
	case "", "line":
		resolution := v.Get("resolution")
		if resolution == "" {
			resolution = "minute"
		}
		if res := dataCompletenessPlot(v.Get("siteID"), v.Get("typeID"), resolution, ts.Line, b); !res.Ok {
			return res
		}
	case "scatter":
		resolution := v.Get("resolution")
		if resolution == "" {
			resolution = "minute"
		}
		if res := dataCompletenessPlot(v.Get("siteID"), v.Get("typeID"), resolution, ts.Scatter, b); !res.Ok {
			return res
		}
	default:
		if res := dataCompletenessSpark(v.Get("siteID"), v.Get("typeID"), b); !res.Ok {
			return res
		}
	}

	return &weft.StatusOK
}

/*
plot draws an svg plot to b.  Assumes f.loadPK has been called first.
*/
func dataCompletenessPlot(siteID, typeID, resolution string, plotter ts.SVGPlot, b *bytes.Buffer) *weft.Result {
	var err error
	// we need the sitePK often so read it once.
	var sitePK int
	if err = dbR.QueryRow(`SELECT sitePK FROM data.site WHERE siteID = $1`,
		siteID).Scan(&sitePK); err != nil {
		if err == sql.ErrNoRows {
			return &weft.NotFound
		}
		return weft.InternalServerError(err)
	}

	var typePK int
	var expected int

	if err = dbR.QueryRow(`SELECT typePK, expected FROM data.completeness_type WHERE typeID = $1`,
		typeID).Scan(&typePK, &expected); err != nil {
		if err == sql.ErrNoRows {
			return &weft.NotFound
		}
		return weft.InternalServerError(err)
	}

	expectedf := float64(expected)
	var p ts.Plot

	var tags []string
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT tag FROM data.completeness_tag JOIN mtr.tag USING (tagpk) WHERE
		sitePK = $1 AND typePK = $2
		ORDER BY tag asc`,
		sitePK, typePK); err != nil {

		return weft.InternalServerError(err)
	}

	defer rows.Close()

	for rows.Next() {
		var s string
		if err = rows.Scan(&s); err != nil {
			return weft.InternalServerError(err)
		}
		tags = append(tags, s)
	}
	rows.Close()

	p.SetSubTitle("Tags: " + strings.Join(tags, ","))
	p.SetTitle(fmt.Sprintf("Site: %s - %s", siteID, strings.Title(typeID)))
	p.SetUnit("completeness")

	switch resolution {
	case "five_minutes":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*2), time.Now().UTC())
		p.SetXLabel("48 hours")

		expectedf /= 288
		rows, err = dbR.Query(`SELECT date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t,
		 sum(count) FROM data.completeness WHERE
		sitePK = $1 AND typePK = $2
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min'
		ORDER BY t ASC`,
			sitePK, typePK)
	case "hour":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*28), time.Now().UTC())
		p.SetXLabel("4 weeks")

		expectedf /= 24
		rows, err = dbR.Query(`SELECT date_trunc('`+resolution+`',time) as t, sum(count) FROM data.completeness WHERE
		sitePK = $1 AND typePK = $2
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time)
		ORDER BY t ASC`,
			sitePK, typePK)
	case "twelve_hours":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*28), time.Now().UTC())
		p.SetXLabel("4 weeks")

		expectedf /= 2
		rows, err = dbR.Query(`SELECT date_trunc('hour', time) + extract(hour from time)::int / 12 * interval '12 hour' as t, sum(count) FROM data.completeness WHERE
		sitePK = $1 AND typePK = $2
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('hour', time) + extract(hour from time)::int / 12 * interval '12 hour'
		ORDER BY t ASC`,
			sitePK, typePK)
	default:
		return weft.BadRequest("invalid resolution")
	}
	if err != nil {
		return weft.InternalServerError(err)
	}

	var pts []ts.Point

	for rows.Next() {
		var pt ts.Point
		var v int
		if err = rows.Scan(&pt.DateTime, &v); err != nil {
			return weft.InternalServerError(err)
		}

		pt.Value = float64(v) / expectedf
		pts = append(pts, pt)
	}

	// Add the latest value to the plot.
	var pt ts.Point

	if err = dbR.QueryRow(`SELECT time, count FROM data.completeness WHERE
			sitePK = $1 AND typePK = $2
			ORDER BY time DESC
			LIMIT 1`,
		sitePK, typePK).Scan(&pt.DateTime, &pt.Value); err != nil {
		// Note: We keep rendering the plot even there's no data.
		if err != sql.ErrNoRows {
			return weft.InternalServerError(err)
		}
		pt.Value = pt.Value / expectedf

		pts = append(pts, pt)
		p.SetLatest(pt, "deepskyblue")
	}

	p.AddSeries(ts.Series{Colour: "deepskyblue", Points: pts})

	if err = plotter.Draw(p, b); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

/*
spark draws an svg spark line to b.  Assumes f.loadPK has been called first.
*/
func dataCompletenessSpark(siteID, typeID string, b *bytes.Buffer) *weft.Result {
	var p ts.Plot

	p.SetXAxis(time.Now().UTC().Add(time.Hour*-12), time.Now().UTC())

	var err error
	var rows *sql.Rows
	var expected int
	var typePK int

	if err = dbR.QueryRow(`SELECT typePK, expected FROM data.completeness_type WHERE typeID = $1`,
		typeID).Scan(&typePK, &expected); err != nil {
		if err == sql.ErrNoRows {
			return &weft.NotFound
		}
		return weft.InternalServerError(err)
	}

	expectedf := float64(expected)

	if rows, err = dbR.Query(`SELECT date_trunc('hour',time) as t, sum(count) FROM data.completeness
		WHERE sitePK = (SELECT sitePK FROM data.site WHERE siteID = $1)
		AND typePK = $2
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('hour',time)
		ORDER BY t ASC`,
		siteID, typePK); err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	var pts []ts.Point

	for rows.Next() {
		var pt ts.Point
		var v int
		if err = rows.Scan(&pt.DateTime, &v); err != nil {
			return weft.InternalServerError(err)
		}
		// No need to scale spark data for display.
		pt.Value = float64(v) / expectedf
		pts = append(pts, pt)
	}

	if len(pts) > 0 {
		p.AddSeries(ts.Series{Colour: "deepskyblue", Points: pts})
	}

	if err = ts.SparkLine.Draw(p, b); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
