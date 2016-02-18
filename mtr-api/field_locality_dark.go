package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/map180/nzmap"
	"net/http"
	"sort"
)

type fieldLocalityDark struct {
}

func (f *fieldLocalityDark) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	var s string

	if err := dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]') 
		FROM (
			SELECT
			localityID AS "LocalityID",
			name AS "Name",
			latitude AS "Latitude",
			longitude AS "Longitude"
			FROM field.locality LEFT JOIN field.metric_latest USING (localityPK) 
			WHERE
			value IS NULL) l`).Scan(&s); err != nil {
		return internalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/json;version=1")

	return &statusOK
}

func (f *fieldLocalityDark) geojsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if result := checkQuery(r, []string{}, []string{}); !result.ok {
		return result
	}

	var s string

	if err := dbR.QueryRow(`SELECT row_to_json(fc)
		FROM ( SELECT 'FeatureCollection' as type, COALESCE(array_to_json(array_agg(f)), '[]') as features
		FROM (SELECT 'Feature' as type,
		ST_AsGeoJSON(q.geom)::json as geometry,
		row_to_json((SELECT l FROM
			(
				SELECT
				localityID AS "localityID",
				name AS "name"
				) as l
				)) as properties FROM field.locality as q LEFT JOIN field.metric_latest USING (localityPK) 
				WHERE
				value IS NULL ) as f ) as fc`).Scan(&s); err != nil {
		return internalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/vnd.geo+json;version=1")

	return &statusOK
}

func (f *fieldLocalityDark) svg(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	var pts nzmap.Points
	var rows *sql.Rows
	var err error

	if rows, err = dbR.Query(`SELECT longitude, latitude
		FROM field.locality LEFT JOIN field.metric_latest USING (localityPK) 
		WHERE
		value IS NULL`); err != nil {
		return internalServerError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var p nzmap.Point

		if err = rows.Scan(&p.Longitude, &p.Latitude); err != nil {
			return internalServerError(err)
		}

		p.Stroke = "magenta"
		p.Fill = "magenta"

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

	h.Set("Content-Type", "image/svg+xml")

	return &statusOK
}
