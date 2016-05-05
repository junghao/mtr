package main

import (
	"database/sql"
	"github.com/lib/pq"
	"net/http"
	"strconv"
)

type dataSite struct {
	siteID              string
	longitude, latitude float64
}

func (d *dataSite) save(r *http.Request) *result {
	if res := checkQuery(r, []string{"siteID", "latitude", "longitude"}, []string{}); !res.ok {
		return res
	}

	d.siteID = r.URL.Query().Get("siteID")

	var err error

	if d.latitude, err = strconv.ParseFloat(r.URL.Query().Get("latitude"), 64); err != nil {
		return badRequest("latitude invalid")
	}

	if d.longitude, err = strconv.ParseFloat(r.URL.Query().Get("longitude"), 64); err != nil {
		return badRequest("longitude invalid")
	}

	// TODO convert to upsert with pg 9.5
	if _, err := db.Exec(`INSERT INTO data.site(siteID, latitude, longitude) VALUES($1, $2, $3)`,
		d.siteID, d.latitude, d.longitude); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			if _, err := db.Exec(`UPDATE data.site SET latitude=$2, longitude=$3 where siteID=$1`,
				d.siteID, d.latitude, d.longitude); err != nil {
				return internalServerError(err)
			}
		} else {
			return internalServerError(err)
		}
	}

	return &statusOK
}

func (d *dataSite) delete(r *http.Request) *result {
	if res := checkQuery(r, []string{"siteID"}, []string{}); !res.ok {
		return res
	}

	d.siteID = r.URL.Query().Get("siteID")

	if _, err := db.Exec(`DELETE FROM data.site where siteID = $1`, d.siteID); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func dataSitePK(siteID string) (int, *result) {
	var pk int

	if err := dbR.QueryRow(`SELECT sitePK FROM data.site where siteID = $1`, siteID).Scan(&pk); err != nil {
		if err == sql.ErrNoRows {
			return pk, badRequest("unknown siteID")
		}
		return pk, internalServerError(err)
	}

	return pk, &statusOK
}
