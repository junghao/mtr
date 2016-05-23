package main

import (
	"github.com/GeoNet/weft"
	"net/http"
	"strconv"
	"time"
)

// put handles http.PUT methods, parsing the request and saving to the db.
func (a *applicationTimer) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"applicationID", "sourceID", "time", "average",
		"count", "fifty", "ninety"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	var err error

	if a.average, err = strconv.Atoi(v.Get("average")); err != nil {
		return weft.BadRequest("invalid average")
	}

	if a.count, err = strconv.Atoi(v.Get("count")); err != nil {
		return weft.BadRequest("invalid count")
	}

	if a.fifty, err = strconv.Atoi(v.Get("fifty")); err != nil {
		return weft.BadRequest("invalid fifty")
	}

	if a.ninety, err = strconv.Atoi(v.Get("ninety")); err != nil {
		return weft.BadRequest("invalid ninety")
	}

	if a.t, err = time.Parse(time.RFC3339, v.Get("time")); err != nil {
		return weft.BadRequest("invalid time")
	}

	if res := a.application.readCreate(v.Get("applicationID")); !res.Ok {
		return res
	}

	if res := a.applicationSource.readCreate(v.Get("sourceID")); !res.Ok {
		return res
	}

	// TODO - what to do when sending from multiple instances and primary key violations?
	if _, err := db.Exec(`INSERT INTO app.timer(applicationPK, sourcePK, time, average, count, fifty, ninety) VALUES($1,$2,$3,$4,$5,$6,$7)`,
		a.application.pk, a.applicationSource.pk, a.t, a.average, a.count, a.fifty, a.ninety); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
