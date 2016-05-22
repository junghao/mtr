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

// create saves the application counter to the db.
// repeat requests at the same time increment the count
// in the db.
func (a *applicationCounter) create() *weft.Result {
	if res := a.applicationType.read(); !res.Ok {
		return res
	}

	if res := a.application.create(); !res.Ok {
		return res
	}

	// TODO convert to UPSERT
	if _, err := db.Exec(`INSERT INTO app.counter(applicationPK, typePK, time, count) VALUES($1,$2,$3,$4)`,
		a.application.pk, a.applicationType.pk, a.t, a.c); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			if _, err := db.Exec(`UPDATE app.counter set count = count + $4
			WHERE applicationPK = $1 AND typePK = $2 AND time = $3`,
				a.application.pk, a.applicationType.pk, a.t, a.c); err != nil {
				return weft.InternalServerError(err)
			}
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

// put handles http.PUT methods, parsing the request and saving to the db.
func (a *applicationCounter) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"applicationID", "typeID", "time", "count"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	var err error

	if a.c, err = strconv.Atoi(v.Get("count")); err != nil {
		return weft.BadRequest("invalid count")
	}

	if a.t, err = time.Parse(time.RFC3339, v.Get("time")); err != nil {
		return weft.BadRequest("invalid time")
	}

	a.applicationType.id = v.Get("typeID")
	a.application.id = v.Get("applicationID")

	return a.create()
}
