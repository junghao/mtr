package main

import (
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
	"strconv"
	"time"
)

type applicationCounter struct {
	application
	applicationType
	t time.Time
	c int
}

func (a *applicationCounter) save(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"applicationID", "typeID", "time", "count"}, []string{}); !res.Ok {
		return res
	}

	var err error

	if a.c, err = strconv.Atoi(r.URL.Query().Get("count")); err != nil {
		return weft.BadRequest("invalid count")
	}

	if a.t, err = time.Parse(time.RFC3339, r.URL.Query().Get("time")); err != nil {
		return weft.BadRequest("invalid time")
	}

	if res := a.applicationType.loadPK(r); !res.Ok {
		return res
	}

	if res := a.application.loadPK(r); !res.Ok {
		return res
	}

	// TODO convert to UPSERT
	if _, err = db.Exec(`INSERT INTO app.counter(applicationPK, typePK, time, count) VALUES($1,$2,$3,$4)`,
		a.applicationPK, a.typePK, a.t, a.c); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			if _, err := db.Exec(`UPDATE app.counter set count = count + $4
			WHERE applicationPK = $1 AND typePK = $2 AND time = $3`,
				a.applicationPK, a.typePK, a.t, a.c); err != nil {
				return weft.InternalServerError(err)
			}
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}
