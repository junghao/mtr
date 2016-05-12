package main

import (
	"database/sql"
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
	"strconv"
)

type dataSite struct {
	sitePK              int
	siteID              string
	longitude, latitude float64
}

func (d *dataSite) save(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "latitude", "longitude"}, []string{}); !res.Ok {
		return res
	}

	d.siteID = r.URL.Query().Get("siteID")

	var err error

	if d.latitude, err = strconv.ParseFloat(r.URL.Query().Get("latitude"), 64); err != nil {
		return weft.BadRequest("latitude invalid")
	}

	if d.longitude, err = strconv.ParseFloat(r.URL.Query().Get("longitude"), 64); err != nil {
		return weft.BadRequest("longitude invalid")
	}

	// TODO convert to upsert with pg 9.5
	if _, err := db.Exec(`INSERT INTO data.site(siteID, latitude, longitude) VALUES($1, $2, $3)`,
		d.siteID, d.latitude, d.longitude); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			if _, err := db.Exec(`UPDATE data.site SET latitude=$2, longitude=$3 where siteID=$1`,
				d.siteID, d.latitude, d.longitude); err != nil {
				return weft.InternalServerError(err)
			}
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func (d *dataSite) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID"}, []string{}); !res.Ok {
		return res
	}

	d.siteID = r.URL.Query().Get("siteID")

	if _, err := db.Exec(`DELETE FROM data.site where siteID = $1`, d.siteID); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (d *dataSite) loadPK(r *http.Request) *weft.Result {
	if err := dbR.QueryRow(`SELECT sitePK FROM data.site where siteID = $1`,
		r.URL.Query().Get("siteID")).Scan(&d.sitePK); err != nil {
		if err == sql.ErrNoRows {
			return weft.BadRequest("unknown siteID")
		}
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
