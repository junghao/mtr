package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/mtr/ts"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var maxAge = time.Duration(-672 * time.Hour)
var future = time.Duration(10 * time.Second)

func fieldMetricHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		var localityID, sourceID, typeID string
		var t time.Time
		var v int
		var err error

		if localityID, sourceID, typeID, err = lst(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if v, err = strconv.Atoi(r.URL.Query().Get("value")); err != nil {
			http.Error(w, "invalid value: "+err.Error(), http.StatusBadRequest)
			return
		}

		if t, err = time.Parse(time.RFC3339, r.URL.Query().Get("time")); err != nil {
			http.Error(w, "invalid time: "+err.Error(), http.StatusBadRequest)
			return
		}

		now := time.Now().UTC()

		if t.Before(now.Add(maxAge)) {
			http.Error(w, "old metric", http.StatusBadRequest)
			return
		}

		if now.Add(future).Before(t) {
			http.Error(w, "future metric", http.StatusBadRequest)
			return
		}

		// Make sure there is not a metric in this hour already
		var f int

		if err = db.QueryRow(`SELECT count(*) FROM field.metric 
			WHERE 
			localityPK = (SELECT localityPK from field.locality WHERE localityID = $1)
			AND 
			sourcePK = (SELECT sourcePK from field.source WHERE sourceID = $2)
			AND
			typePK = (SELECT typePK from field.type where typeID = $3)
			AND 
			date_trunc('hour', time) = $4`, localityID, sourceID, typeID, t.Truncate(time.Hour)).Scan(&f); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if f != 0 {
			http.Error(w, "metric exists for this hour already", http.StatusBadRequest)
			return
		}

		// Insert the metric
		var c sql.Result
		if c, err = db.Exec(`INSERT INTO field.metric(localityPK, sourcePK, typePK, time, value) 
			select localityPK, sourcePK, typePK, $4, $5 
			FROM field.locality, field.source, field.type 
			WHERE 
			localityID = $1
			AND 
			sourceID = $2 
			AND
			typeID = $3`,
			localityID, sourceID, typeID, t, int32(v)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var i int64
		i, err = c.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if i == 0 {
			http.Error(w, "no data inserted check *ID parameters are valid.", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	case "DELETE":
		var localityID, sourceID, typeID string
		var err error

		if localityID, sourceID, typeID, err = lst(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, err = db.Exec(`DELETE FROM field.metric USING field.locality, field.source, field.type
			WHERE metric.localityPK = locality.localityPK 
			AND metric.sourcePK = source.sourcePK 
			AND metric.typePK = type.typePK 
			AND locality.localityID = $1
			AND source.sourceID = $2
			AND type.typeID = $3`, localityID, sourceID, typeID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	case "GET":
		var localityID, sourceID, typeID string
		var err error

		if localityID, sourceID, typeID, err = lst(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var b bytes.Buffer

		switch r.Header.Get("Accept") {
		case "text/csv":
			w.Header().Set("Content-Disposition", `attachment; filename="MTR-`+strings.Replace(localityID+`-`+sourceID+`-`+typeID, " ", "-", -1)+`.csv"`)
			w.Header().Set("Content-Type", "text/csv")
			err = metricCSV(localityID, sourceID, typeID, &b)
		default:
			w.Header().Set("Content-Type", "image/svg+xml")
			switch r.URL.Query().Get("plot") {
			case "spark":
				err = metricSparkSVG(localityID, sourceID, typeID, &b)
			case "":
				err = metricSVG(localityID, sourceID, typeID, &b)
			default:
				http.Error(w, "bad plot parameter", http.StatusBadRequest)
				return
			}
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		b.WriteTo(w)

	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

}

func metricCSV(localityID, sourceID, typeID string, b *bytes.Buffer) (err error) {
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT format('%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value) as csv FROM field.metric 
			WHERE 
			localityPK = (SELECT localityPK from field.locality WHERE localityID = $1)
			AND 
			sourcePK = (SELECT sourcePK from field.source WHERE sourceID = $2)
			AND
			typePK = (SELECT typePK from field.type where typeID = $3)
			ORDER BY time ASC`,
		localityID, sourceID, typeID); err != nil {
		return
	}
	defer rows.Close()

	var d string

	b.Write([]byte("date-time," + typeID))
	b.Write(eol)
	for rows.Next() {
		err = rows.Scan(&d)
		if err != nil {
			return
		}
		b.Write([]byte(d))
		b.Write(eol)
	}
	rows.Close()

	return
}

func metricSVG(localityID, sourceID, typeID string, b *bytes.Buffer) (err error) {
	var p ts.Plot

	if p, err = plot(localityID, sourceID, typeID); err != nil {
		return
	}

	return ts.Scatter.Draw(p, b)
}

func metricSparkSVG(localityID, sourceID, typeID string, b *bytes.Buffer) (err error) {
	var p ts.Plot

	if p, err = plot(localityID, sourceID, typeID); err != nil {
		return
	}

	return ts.SparkScatterLatest.Draw(p, b)
}

func plot(localityID, sourceID, typeID string) (p ts.Plot, err error) {
	var name string
	err = db.QueryRow(`SELECT name FROM field.locality WHERE localityID = $1`, localityID).Scan(&name)
	if err != nil {
		return
	}

	p.SetTitle(fmt.Sprintf("%s - %s - %s", name, sourceID, strings.Title(typeID)))

	var min, max int32

	// ignore errors, there might not be a min,max and next DB query will catch a failed db
	dbR.QueryRow(`SELECT min,max FROM field.threshold
			WHERE
			localityPK = (SELECT localityPK from field.locality WHERE localityID = $1)
			AND
			sourcePK = (SELECT sourcePK from field.source WHERE sourceID = $2)
			AND
			typePK = (SELECT typePK from field.type where typeID = $3)`,
		localityID, sourceID, typeID).Scan(&min, &max)

	if min == 0 && max == 0 {
		min = math.MinInt32
		max = math.MaxInt32
	} else {
		switch typeID {
		case "voltage":
			p.SetThreshold(float64(min)*0.001, float64(max)*0.001)
		default:
			p.SetThreshold(float64(min), float64(max))
		}
	}

	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT time, value FROM field.metric
			WHERE
			localityPK = (SELECT localityPK from field.locality WHERE localityID = $1)
			AND
			sourcePK = (SELECT sourcePK from field.source WHERE sourceID = $2)
			AND
			typePK = (SELECT typePK from field.type where typeID = $3)
			ORDER BY time ASC`,
		localityID, sourceID, typeID); err != nil {
		return
	}
	defer rows.Close()

	var t time.Time
	var v int32
	var pts []ts.Point

	switch typeID {
	case "voltage":
		p.SetUnit("V")
		p.SetYLabel("Voltage (V)")
		p.SetYAxis(0.0, 25.0)
		for rows.Next() {
			err = rows.Scan(&t, &v)
			if err != nil {
				return
			}
			if v < min || v > max {
				pts = append(pts, ts.Point{DateTime: t, Value: float64(v) * 0.001, Colour: "firebrick"})
			} else {
				pts = append(pts, ts.Point{DateTime: t, Value: float64(v) * 0.001, Colour: "darkcyan"})
			}
		}
	default:
		for rows.Next() {
			err = rows.Scan(&t, &v)
			if err != nil {
				return
			}
			if v < min || v > max {
				pts = append(pts, ts.Point{DateTime: t, Value: float64(v), Colour: "firebrick"})
			} else {
				pts = append(pts, ts.Point{DateTime: t, Value: float64(v), Colour: "darkcyan"})
			}
		}
	}
	rows.Close()

	p.AddSeries(ts.Series{Label: sourceID, Points: pts})
	p.SetXAxis(time.Now().UTC().Add(maxAge), time.Now().UTC())

	return
}

func lst(r *http.Request) (localityID, sourceID, typeID string, err error) {

	if localityID = r.URL.Query().Get("localityID"); localityID == "" {
		err = fmt.Errorf("localityID is a required parameter")
		return
	}

	if sourceID = r.URL.Query().Get("sourceID"); sourceID == "" {
		err = fmt.Errorf("sourceID is a required parameter")
		return
	}

	if typeID = r.URL.Query().Get("typeID"); typeID == "" {
		err = fmt.Errorf("typeID is a required paramter")
		return
	}

	return
}
