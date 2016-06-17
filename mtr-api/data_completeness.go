package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/mtr/ts"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"github.com/lib/pq"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type dataCompleteness struct{}

func (a dataCompleteness) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID", "time", "count"}, []string{}); !res.Ok {
		return res
	}

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

	return &weft.StatusOK
}

func (a dataCompleteness) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	siteID := v.Get("siteID")
	typeID := v.Get("typeID")

	if _, err := db.Exec(`DELETE FROM data.completeness where
				sitePK = (SELECT sitePK FROM data.site WHERE siteID = $1)
				AND typePK = (SELECT typePK FROM data.type WHERE typeID = $2)`, siteID, typeID); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (a dataCompleteness) proto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{}); !res.Ok {
		return res
	}

	typeID := r.URL.Query().Get("typeID")
	siteID := r.URL.Query().Get("siteID")

	var err error
	var rows *sql.Rows
	var typePK int
	var expected int

	if err = dbR.QueryRow(`SELECT typePK, expected FROM data.completeness_type WHERE typeID = $1`,
		typeID).Scan(&typePK, &expected); err != nil {
		if err == sql.ErrNoRows {
			return &weft.NotFound
		}
		return weft.InternalServerError(err)
	}

	expectedf := float32(expected) / 288

	rows, err = dbR.Query(`SELECT date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t,
		 sum(count) FROM data.completeness WHERE
		sitePK = (SELECT sitePK FROM data.site WHERE siteID = $1)
		AND typePK = $2
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min'
		ORDER BY t ASC`,
		siteID, typePK)

	if err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var dcr mtrpb.DataCompletenessResult

	for rows.Next() {
		var count int

		if err = rows.Scan(&t, &count); err != nil {
			return weft.InternalServerError(err)
		}

		c := float32(count) / expectedf
		dc := mtrpb.DataCompleteness{TypeID: typeID, SiteID: siteID, Completeness: c, Seconds: t.Unix()}
		dcr.Result = append(dcr.Result, &dc)
	}
	rows.Close()

	var by []byte

	if by, err = proto.Marshal(&dcr); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	h.Set("Content-Type", "application/x-protobuf")

	return &weft.StatusOK
}

func (a dataCompleteness) svg(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{"plot", "resolution", "yrange"}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	switch r.URL.Query().Get("plot") {
	case "", "line":
		resolution := v.Get("resolution")
		if resolution == "" {
			resolution = "minute"
		}
		if res := a.plot(v.Get("siteID"), v.Get("typeID"), resolution, b); !res.Ok {
			return res
		}
	default:
		if res := a.spark(v.Get("siteID"), v.Get("typeID"), b); !res.Ok {
			return res
		}
	}

	h.Set("Content-Type", "image/svg+xml")

	return &weft.StatusOK
}

/*
plot draws an svg plot to b.  Assumes f.loadPK has been called first.
*/
func (a dataCompleteness) plot(siteID, typeID, resolution string, b *bytes.Buffer) *weft.Result {
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

	defer rows.Close()

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
	rows.Close()

	// Add the latest value to the plot - this may be different to the average at minute or hour resolution.
	var pt ts.Point

	if err = dbR.QueryRow(`SELECT time, count FROM data.completeness WHERE
			sitePK = $1 AND typePK = $2
			ORDER BY time DESC
			LIMIT 1`,
		sitePK, typePK).Scan(&pt.DateTime, &pt.Value); err != nil {
		return weft.InternalServerError(err)
	}

	pt.Value = pt.Value / expectedf

	pts = append(pts, pt)
	p.SetLatest(pt, "deepskyblue")

	p.AddSeries(ts.Series{Colour: "deepskyblue", Points: pts})

	if err = ts.Line.Draw(p, b); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

/*
spark draws an svg spark line to b.  Assumes f.loadPK has been called first.
*/
func (a dataCompleteness) spark(siteID, typeID string, b *bytes.Buffer) *weft.Result {
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
	rows.Close()

	p.AddSeries(ts.Series{Colour: "deepskyblue", Points: pts})

	if err = ts.SparkLine.Draw(p, b); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
