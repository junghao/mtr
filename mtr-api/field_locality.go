package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/map180/nzmap"
	"github.com/lib/pq"
	"net/http"
	"strconv"
)

type fieldLocality struct {
	localityID, name    string
	longitude, latitude float64
	localityPK          int
}

func (f *fieldLocality) save(r *http.Request) *result {
	if res := checkQuery(r, []string{"localityID", "name", "longitude", "latitude"}, []string{}); !res.ok {
		return res
	}

	f.localityID = r.URL.Query().Get("localityID")
	f.name = r.URL.Query().Get("name")

	var err error

	if f.latitude, err = strconv.ParseFloat(r.URL.Query().Get("latitude"), 64); err != nil {
		return badRequest("latitude invalid")
	}

	if f.longitude, err = strconv.ParseFloat(r.URL.Query().Get("longitude"), 64); err != nil {
		return badRequest("longitude invalid")
	}

	if _, err = db.Exec(`INSERT INTO field.locality(localityID, name, latitude, longitude) VALUES($1,$2,$3,$4)`,
		f.localityID, f.name, f.latitude, f.longitude); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == `23505` {
			// ignore unique errors
		} else {
			return internalServerError(err)
		}
	}

	if _, err = db.Exec(`UPDATE field.locality SET name=$2, latitude=$3, longitude=$4 WHERE localityID=$1`,
		f.localityID, f.name, f.latitude, f.longitude); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func (f *fieldLocality) delete(r *http.Request) *result {
	if res := checkQuery(r, []string{"localityID"}, []string{}); !res.ok {
		return res
	}

	f.localityID = r.URL.Query().Get("localityID")

	if _, err := db.Exec(`DELETE FROM field.locality WHERE localityID = $1`, f.localityID); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func (f *fieldLocality) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{"localityID"}); !res.ok {
		return res
	}

	f.localityID = r.URL.Query().Get("localityID")

	var s string
	var err error

	switch f.localityID {
	case "":
		err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]') 
		FROM (
			SELECT
			localityID AS "LocalityID",
			name AS "Name",
			latitude AS "Latitude",
			longitude AS "Longitude"
			FROM field.locality
			) l`).Scan(&s)
	default:
		err = dbR.QueryRow(`SELECT row_to_json(l) 
		FROM (
			SELECT
			localityID AS "LocalityID",
			name AS "Name",
			latitude AS "Latitude",
			longitude AS "Longitude"
			FROM field.locality WHERE localityID=$1
			) l`, f.localityID).Scan(&s)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return &notFound
		}
		return internalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/json;version=1")

	return &statusOK
}

func (f *fieldLocality) geojsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{"localityID"}); !res.ok {
		return res
	}

	f.localityID = r.URL.Query().Get("localityID")

	var s string
	var err error

	switch f.localityID {
	case "":
		err = dbR.QueryRow(`SELECT row_to_json(fc)
		FROM ( SELECT 'FeatureCollection' as type, COALESCE(array_to_json(array_agg(f)), '[]') as features
		FROM (SELECT 'Feature' as type,
		ST_AsGeoJSON(q.geom)::json as geometry,
		row_to_json((SELECT l FROM
			(
				SELECT
				localityID AS "localityID",
				name AS "name"
				) as l
				)) as properties FROM field.locality as q) as f ) as fc`).Scan(&s)
	default:
		err = dbR.QueryRow(`SELECT row_to_json(f)
		FROM (SELECT 'Feature' as type,
		ST_AsGeoJSON(q.geom)::json as geometry,
		row_to_json((SELECT l FROM
			(
				SELECT
				localityID AS "localityID",
				name AS "name"
				) as l
				)) as properties FROM field.locality as q WHERE localityID=$1) as f`, f.localityID).Scan(&s)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return &notFound
		}
		return internalServerError(err)
	}

	b.WriteString(s)

	h.Set("Content-Type", "application/vnd.geo+json;version=1")

	return &statusOK
}

func (f *fieldLocality) svg(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"localityID"}, []string{}); !res.ok {
		return res
	}

	f.localityID = r.URL.Query().Get("localityID")

	var p nzmap.Point

	if err := dbR.QueryRow(`SELECT longitude, latitude FROM field.locality WHERE localityID = $1`,
		f.localityID).Scan(&p.Longitude, &p.Latitude); err != nil {
		if err == sql.ErrNoRows {
			return &notFound
		}
		return internalServerError(err)
	}

	p.Icon(b)

	if p.Visible() {
		b.WriteString(fmt.Sprintf("<path d=\"M%d %d l5 0 l-5 -8 l-5 8 Z\" stroke-width=\"0\" fill=\"blue\" opacity=\"0.7\"></path>", p.X(), p.Y()))
	}

	b.WriteString("</svg>")

	h.Set("Content-Type", "image/svg+xml")

	return &statusOK
}
