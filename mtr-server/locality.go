package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/map180/nzmap"
	"net/http"
	"sort"
	"strconv"
)

func localityHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		var localityID, name string
		var latitude, longitude float64
		var err error

		if localityID = r.URL.Query().Get("localityID"); localityID == "" {
			http.Error(w, "localityID is a required parameter", http.StatusBadRequest)
			return
		}

		if name = r.URL.Query().Get("name"); name == "" {
			http.Error(w, "name is a required parameter", http.StatusBadRequest)
			return
		}

		if latitude, err = strconv.ParseFloat(r.URL.Query().Get("latitude"), 64); err != nil {
			http.Error(w, "latitude invalid: "+err.Error(), http.StatusBadRequest)
			return
		}

		if longitude, err = strconv.ParseFloat(r.URL.Query().Get("longitude"), 64); err != nil {
			http.Error(w, "longitude invalid: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Ignore errors - will catch any on the update.
		db.Exec(`INSERT INTO field.locality(localityID, name, latitude, longitude) VALUES($1,$2,$3,$4)`,
			localityID, name, latitude, longitude)

		var c sql.Result
		if c, err = db.Exec(`UPDATE field.locality SET name=$2, latitude=$3, longitude=$4 WHERE localityID=$1`,
			localityID, name, latitude, longitude); err != nil {
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
			http.Error(w, "no data inserted check localityID is valid.", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	case "DELETE":
		var localityID string

		if localityID = r.URL.Query().Get("localityID"); localityID == "" {
			http.Error(w, "localityID is a required parameter", http.StatusBadRequest)
			return
		}

		if _, err := db.Exec(`DELETE FROM field.locality WHERE localityID = $1`, localityID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	case "GET":
		var err error
		var b bytes.Buffer

		localityID := r.URL.Query().Get("localityID") // localityID is optional

		switch r.Header.Get("Accept") {
		case "application/vnd.geo+json;version=1":
			w.Header().Set("Content-Type", "application/vnd.geo+json;version=1")
			switch localityID {
			case "":
				err = localitiesGeoJSONV1(&b)
			default:
				err = localityGeoJSONV1(localityID, &b)
			}
		case "application/json;version=1":
			w.Header().Set("Content-Type", "application/json;version=1")
			switch localityID {
			case "":
				err = localitiesJSONV1(&b)
			default:
				err = localityJSONV1(localityID, &b)
			}
		default:
			w.Header().Set("Content-Type", "image/svg+xml")
			switch localityID {
			case "":
				http.Error(w, "page not found", http.StatusNotFound)
			default:
				err = localitySVG(localityID, &b)
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

func localityDarkHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var err error
		var b bytes.Buffer

		switch r.Header.Get("Accept") {
		case "application/json;version=1":
			w.Header().Set("Content-Type", "application/json;version=1")
			err = localitiesDarkJSONV1(&b)
		case "application/vnd.geo+json;version=1":
			w.Header().Set("Content-Type", "application/vnd.geo+json;version=1")
			err = localitiesDarkGeoJSONV1(&b)
		default:
			w.Header().Set("Content-Type", "image/svg+xml")
			err = localitiesDarkSVG(&b)
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

func localitiesJSONV1(b *bytes.Buffer) (err error) {
	var d string
	if err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]') 
		FROM (
			SELECT
			localityID AS "LocalityID",
			name AS "Name",
			latitude AS "Latitude",
			longitude AS "Longitude"
			FROM field.locality
			) l`).Scan(&d); err != nil {
		return
	}

	b.WriteString(d)

	return
}

func localitiesGeoJSONV1(b *bytes.Buffer) (err error) {
	var s string

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

	b.WriteString(s)

	return
}

func localityJSONV1(localityID string, b *bytes.Buffer) (err error) {
	var s string

	err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]') 
		FROM (
			SELECT
			localityID AS "LocalityID",
			name AS "Name",
			latitude AS "Latitude",
			longitude AS "Longitude"
			FROM field.locality WHERE localityID=$1
			) l`, localityID).Scan(&s)

	b.WriteString(s)

	return
}

func localityGeoJSONV1(localityID string, b *bytes.Buffer) (err error) {
	var s string

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
	)) as properties FROM field.locality as q WHERE localityID=$1) as f ) as fc`, localityID).Scan(&s)

	b.WriteString(s)

	return
}

func localitySVG(localityID string, b *bytes.Buffer) (err error) {
	var p nzmap.Point

	// return an empty map if no locality is found.
	_ = dbR.QueryRow(`SELECT longitude, latitude FROM field.locality WHERE localityID = $1`,
		localityID).Scan(&p.Longitude, &p.Latitude)

	p.Icon(b)
	if p.Visible() {
		b.WriteString(fmt.Sprintf("<path d=\"M%d %d l5 0 l-5 -8 l-5 8 Z\" stroke-width=\"0\" fill=\"blue\" opacity=\"0.7\"></path>", p.X(), p.Y()))
	}
	b.WriteString("</svg>")

	return
}

func localitiesDarkGeoJSONV1(b *bytes.Buffer) (err error) {
	var s string

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
	)) as properties FROM field.locality as q LEFT JOIN field.metric_summary USING (localityPK) 
			WHERE
			value IS NULL ) as f ) as fc`).Scan(&s)

	b.WriteString(s)

	return
}

func localitiesDarkJSONV1(b *bytes.Buffer) (err error) {
	var s string

	err = dbR.QueryRow(`SELECT COALESCE(array_to_json(array_agg(row_to_json(l))), '[]') 
		FROM (
			SELECT
			localityID AS "LocalityID",
			name AS "Name",
			latitude AS "Latitude",
			longitude AS "Longitude"
			FROM field.locality LEFT JOIN field.metric_summary USING (localityPK) 
			WHERE
			value IS NULL) l`).Scan(&s)

	b.WriteString(s)

	return
}

/*
SVG map of localities that have no metrics - they have gone dark.
*/
func localitiesDarkSVG(b *bytes.Buffer) (err error) {
	var pts nzmap.Points
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT longitude, latitude
		FROM field.locality LEFT JOIN field.metric_summary USING (localityPK) 
		WHERE
		value IS NULL`); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var p nzmap.Point

		if err = rows.Scan(&p.Longitude, &p.Latitude); err != nil {
			return
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

	return

}
