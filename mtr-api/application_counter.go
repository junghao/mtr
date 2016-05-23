package main

import (
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
	"strconv"
	"time"
)


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

	if res := a.applicationType.read(v.Get("typeID")); !res.Ok {
		return res
	}

	if res := a.application.readCreate(v.Get("applicationID")); !res.Ok {
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
