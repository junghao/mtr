package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/map180"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type fieldLatest struct {
	typeID string
}

func (f *fieldLatest) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{"typeID"}); !res.ok {
		return res
	}

	f.typeID = r.URL.Query().Get("typeID")

	var s string
	var err error

	switch f.typeID {
	case "":
		err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]') 
		FROM (
			SELECT localityID AS "LocalityID", name as "Name", latitude AS "Latitude", longitude AS "Longitude", 
			sourceID AS "SourceID", time AS "Time", value AS "Value",
			typeID AS "TypeID", 
			unit AS "Unit",
			CASE WHEN threshold.lower is NULL THEN 0 ELSE threshold.lower END AS "Lower",  
			CASE WHEN threshold.upper is NULL THEN 0 ELSE threshold.upper END AS "Upper"
			FROM field.metric_latest LEFT OUTER JOIN field.threshold USING (localityPK, sourcePK, typePK)
			JOIN field.locality USING (localityPK) 
			JOIN field.source USING (sourcepk) 
			JOIN field.type USING (typepk)
			) l`).Scan(&s)
	default:
		err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]') 
		FROM (
			SELECT localityID AS "LocalityID", name as "Name", latitude AS "Latitude", longitude AS "Longitude", 
			sourceID AS "SourceID", time AS "Time", value AS "Value",
			typeID AS "TypeID", 
			unit AS "Unit",
			CASE WHEN threshold.lower is NULL THEN 0 ELSE threshold.lower END AS "Lower",  
			CASE WHEN threshold.upper is NULL THEN 0 ELSE threshold.upper END AS "Upper"
			FROM field.metric_latest LEFT OUTER JOIN field.threshold USING (localityPK, sourcePK, typePK)
			JOIN field.locality USING (localityPK) 
			JOIN field.source USING (sourcepk) 
			JOIN field.type USING (typepk)
			WHERE typepk= (select typePK from field.type where typeID = $1)) l`, f.typeID).Scan(&s)
	}
	if err != nil {
		return internalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/json;version=1")

	return &statusOK
}

func (f *fieldLatest) geojsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{"typeID"}); !res.ok {
		return res
	}

	f.typeID = r.URL.Query().Get("typeID")

	var s string
	var err error

	switch f.typeID {
	case "":
		err = dbR.QueryRow(`SELECT row_to_json(fc)
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
						CASE WHEN threshold.lower is NULL THEN 0 ELSE threshold.lower END AS "lower",  
						CASE WHEN threshold.upper is NULL THEN 0 ELSE threshold.upper END AS "upper"
						) as l
	)) as properties FROM field.metric_latest LEFT OUTER JOIN field.threshold USING (localityPK, sourcePK, typePK)
	JOIN field.locality as q USING (localityPK) 
	JOIN field.source USING (sourcepk) 
	JOIN field.type USING (typepk)) as f ) as fc`).Scan(&s)
	default:
		err = dbR.QueryRow(`SELECT row_to_json(fc)
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
						CASE WHEN threshold.lower is NULL THEN 0 ELSE threshold.lower END AS "lower",  
						CASE WHEN threshold.upper is NULL THEN 0 ELSE threshold.upper END AS "upper"
						) as l
	)) as properties FROM field.metric_latest LEFT OUTER JOIN field.threshold USING (localityPK, sourcePK, typePK)
	JOIN field.locality as q USING (localityPK) 
	JOIN field.source USING (sourcepk) 
	JOIN field.type USING (typepk)
	WHERE type.typePK = (select typePK from field.type where typeID = $1)) as f ) as fc`, f.typeID).Scan(&s)
	}
	if err != nil {
		return internalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/vnd.geo+json;version=1")

	return &statusOK
}

func (f *fieldLatest) svg(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"bbox", "width"}, []string{"typeID", "insetBox"}); !res.ok {
		return res
	}

	var pts map180.Points
	var rows *sql.Rows
	var width int
	var err error

	f.typeID = r.URL.Query().Get("typeID")
	bbox := r.URL.Query().Get("bbox")

	if err = map180.ValidBbox(bbox); err != nil {
		return badRequest(err.Error())
	}

	if width, err = strconv.Atoi(r.URL.Query().Get("width")); err != nil {
		return badRequest("invalid width")
	}

	var insetBbox string

	if r.URL.Query().Get("insetBbox") != "" {
		insetBbox = r.URL.Query().Get("insetBbox")

		if err = map180.ValidBbox(insetBbox); err != nil {
			return badRequest(err.Error())
		}
	}

	switch f.typeID {
	case "":
		rows, err = dbR.Query(`SELECT longitude, latitude, time, value,
			CASE WHEN threshold.lower is NULL THEN 0 ELSE threshold.lower END AS "lower",
			CASE WHEN threshold.upper is NULL THEN 0 ELSE threshold.upper END AS "upper"
			FROM field.metric_latest LEFT OUTER JOIN field.threshold USING (localityPK, sourcePK, typePK)
			JOIN field.locality USING (localityPK)`)
	default:
		rows, err = dbR.Query(`SELECT longitude, latitude, time, value,
			CASE WHEN threshold.lower is NULL THEN 0 ELSE threshold.lower END AS "lower",
			CASE WHEN threshold.upper is NULL THEN 0 ELSE threshold.upper END AS "upper"
			FROM field.metric_latest LEFT OUTER JOIN field.threshold USING (localityPK, sourcePK, typePK)
			JOIN field.locality USING (localityPK) 
			WHERE typePK = (SELECT typePK FROM field.type WHERE typeID = $1)`, f.typeID)
	}
	if err != nil {
		return internalServerError(err)
	}
	defer rows.Close()

	ago := time.Now().UTC().Add(time.Hour * -48)

	for rows.Next() {
		var p map180.Point
		var t time.Time
		var min, max, v int

		if err = rows.Scan(&p.Longitude, &p.Latitude, &t, &v, &min, &max); err != nil {
			return internalServerError(err)
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

	if err = wm.Map(bbox, width, pts, insetBbox, b); err != nil {
		return internalServerError(err)
	}

	for _, p := range pts {
		b.WriteString(fmt.Sprintf("<path d=\"M%d %d l5 0 l-5 -10 l-5 10 Z\" stroke-width=\"2\" fill=\"%s\" stroke=\"%s\" opacity=\"0.9\"></path>",
			p.X(), p.Y(), p.Fill, p.Stroke))
	}

	b.WriteString("</svg>")

	h.Set("Content-Type", "image/svg+xml")

	return &statusOK
}
