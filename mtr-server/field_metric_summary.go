package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/map180/nzmap"
	"net/http"
	"sort"
	"time"
)

func fieldMetricSummaryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var err error
		var b bytes.Buffer

		typeID := r.URL.Query().Get("typeID") // optional

		switch r.Header.Get("Accept") {
		case "application/json;version=1":
			w.Header().Set("Content-Type", "application/json;version=1")
			switch typeID {
			case "":
				err = fmsSummaryJSONV1(&b)
			default:
				err = fmSummaryJSONV1(typeID, &b)

			}
		case "application/vnd.geo+json;version=1":
			w.Header().Set("Content-Type", "application/vnd.geo+json;version=1")
			switch typeID {
			case "":
				err = fmsSummaryGeoJSONV1(&b)
			default:
				err = fmSummaryGeoJSONV1(typeID, &b)
			}
		default:
			w.Header().Set("Content-Type", "image/svg+xml")
			err = fieldMetricSummarySVG(typeID, &b)
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

/*
Simple JSON summary of all metric types
*/
func fmsSummaryJSONV1(b *bytes.Buffer) (err error) {
	var d string
	if err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]') 
		FROM (
			SELECT localityID AS "LocalityID", name as "Name", latitude AS "Latitude", longitude AS "Longitude", 
			sourceID AS "SourceID", time AS "Time", value AS "Value",
			typeID AS "TypeID", 
			unit AS "Unit",
			CASE WHEN threshold.min is NULL THEN 0 ELSE threshold.min END AS "Min",  
			CASE WHEN threshold.max is NULL THEN 0 ELSE threshold.max END AS "Max"
			FROM field.metric_summary LEFT OUTER JOIN field.threshold USING (localityPK, sourcePK, typePK)
			JOIN field.locality USING (localityPK) 
			JOIN field.source USING (sourcepk) 
			JOIN field.type USING (typepk)
			) l`).Scan(&d); err != nil {
		return
	}

	b.WriteString(d)

	return
}

/*
Simple JSON summary of a single metric type.
*/
func fmSummaryJSONV1(typeID string, b *bytes.Buffer) (err error) {
	var d string
	if err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]') 
		FROM (
			SELECT localityID AS "LocalityID", name as "Name", latitude AS "Latitude", longitude AS "Longitude", 
			sourceID AS "SourceID", time AS "Time", value AS "Value",
			typeID AS "TypeID", 
			unit AS "Unit",
			CASE WHEN threshold.min is NULL THEN 0 ELSE threshold.min END AS "Min",  
			CASE WHEN threshold.max is NULL THEN 0 ELSE threshold.max END AS "Max"
			FROM field.metric_summary LEFT OUTER JOIN field.threshold USING (localityPK, sourcePK, typePK)
			JOIN field.locality USING (localityPK) 
			JOIN field.source USING (sourcepk) 
			JOIN field.type USING (typepk)
			WHERE typepk= (select typePK from field.type where typeID = $1)) l`, typeID).Scan(&d); err != nil {
		return
	}

	b.WriteString(d)

	return
}

/*
GeoJSON summary of a single metric type.
*/
func fmSummaryGeoJSONV1(typeID string, b *bytes.Buffer) (err error) {
	var d string
	if err = dbR.QueryRow(`SELECT row_to_json(fc)
		FROM ( SELECT 'FeatureCollection' as type, COALESCE(array_to_json(array_agg(f)), '[]') as features
			FROM (SELECT 'Feature' as type,
				ST_AsGeoJSON(q.geom)::json as geometry,
				row_to_json((SELECT l FROM
					(
						SELECT
						localityID AS "localityID", name as "name",
						sourceID AS "sourceID", time AS "time", value AS "value",
						typeID AS "typeID", 
						unit AS "unit",
						CASE WHEN threshold.min is NULL THEN 0 ELSE threshold.min END AS "min",  
						CASE WHEN threshold.max is NULL THEN 0 ELSE threshold.max END AS "max"
						) as l
	)) as properties FROM field.metric_summary LEFT OUTER JOIN field.threshold USING (localityPK, sourcePK, typePK)
	JOIN field.locality as q USING (localityPK) 
	JOIN field.source USING (sourcepk) 
	JOIN field.type USING (typepk)
	WHERE type.typePK = (select typePK from field.type where typeID = $1)) as f ) as fc`, typeID).Scan(&d); err != nil {
		return
	}

	b.WriteString(d)

	return
}

/*
GeoJSON summary of all metrics.
*/
func fmsSummaryGeoJSONV1(b *bytes.Buffer) (err error) {
	var d string
	if err = dbR.QueryRow(`SELECT row_to_json(fc)
		FROM ( SELECT 'FeatureCollection' as type, COALESCE(array_to_json(array_agg(f)), '[]') as features
			FROM (SELECT 'Feature' as type,
				ST_AsGeoJSON(q.geom)::json as geometry,
				row_to_json((SELECT l FROM
					(
						SELECT
						localityID AS "localityID", name as "name",
						sourceID AS "sourceID", time AS "time", value AS "value",
						typeID AS "typeID", 
						unit AS "unit",
						CASE WHEN threshold.min is NULL THEN 0 ELSE threshold.min END AS "min",  
						CASE WHEN threshold.max is NULL THEN 0 ELSE threshold.max END AS "max"
						) as l
	)) as properties FROM field.metric_summary LEFT OUTER JOIN field.threshold USING (localityPK, sourcePK, typePK)
	JOIN field.locality as q USING (localityPK) 
	JOIN field.source USING (sourcepk) 
	JOIN field.type USING (typepk)) as f ) as fc`).Scan(&d); err != nil {
		return
	}

	b.WriteString(d)

	return
}

/*
SVG map of metrics.  Leave typeID zero for all metrics.
*/
func fieldMetricSummarySVG(typeID string, b *bytes.Buffer) (err error) {
	var pts nzmap.Points
	var rows *sql.Rows

	switch typeID {
	case "":
		rows, err = dbR.Query(`SELECT longitude, latitude, time, value,
			CASE WHEN threshold.min is NULL THEN 0 ELSE threshold.min END AS "min",
			CASE WHEN threshold.max is NULL THEN 0 ELSE threshold.max END AS "max"
			FROM field.metric_summary LEFT OUTER JOIN field.threshold USING (localityPK, sourcePK, typePK)
			JOIN field.locality USING (localityPK)`)
	default:
		rows, err = dbR.Query(`SELECT longitude, latitude, time, value,
			CASE WHEN threshold.min is NULL THEN 0 ELSE threshold.min END AS "min",
			CASE WHEN threshold.max is NULL THEN 0 ELSE threshold.max END AS "max"
			FROM field.metric_summary LEFT OUTER JOIN field.threshold USING (localityPK, sourcePK, typePK)
			JOIN field.locality USING (localityPK) 
			WHERE typePK = (SELECT typePK FROM field.type WHERE typeID = $1)`, typeID)
	}
	if err != nil {
		return
	}
	defer rows.Close()

	ago := time.Now().UTC().Add(time.Hour * -48)

	for rows.Next() {
		var p nzmap.Point
		var t time.Time
		var min, max, v int

		if err = rows.Scan(&p.Longitude, &p.Latitude, &t, &v, &min, &max); err != nil {
			return
		}
		switch {
		case min == 0 && max == 0:
			p.Fill = "lightskyblue"
			p.Stroke = "lightskyblue"
			p.Value = 3.0
		case v < min || v > max:
			p.Fill = "crimson"
			p.Stroke = "crimson"
			p.Value = 1.0
		default:
			p.Fill = "lawngreen"
			p.Stroke = "lawngreen"
			p.Value = 4.0
		}

		// Add a border if the metric is old
		if t.Before(ago) {
			p.Stroke = "magenta"
			p.Value = 2.0
		}

		pts = append(pts, p)
	}
	rows.Close()

	sort.Sort(pts)

	pts.Medium(b)

	for _, p := range pts {
		if p.Visible() {
			b.WriteString(fmt.Sprintf("<path d=\"M%d %d l5 0 l-5 -10 l-5 10 Z\" stroke-width=\"2\" fill=\"%s\" stroke=\"%s\" opacity=\"0.9\"></path>",
				p.X(), p.Y(), p.Fill, p.Stroke))
		}
	}

	b.WriteString("</svg>")

	return

}
