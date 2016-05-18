package main

import (

	"github.com/GeoNet/weft"
	"net/http"
	"strconv"
	"time"
)

type applicationTimer struct {
	application
	applicationSource
	t                             time.Time
	average, count, fifty, ninety int
}

func (a *applicationTimer) save(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"applicationID", "sourceID", "time", "average",
		"count", "fifty", "ninety"}, []string{}); !res.Ok {
		return res
	}

	var err error

	if a.average, err = strconv.Atoi(r.URL.Query().Get("average")); err != nil {
		return weft.BadRequest("invalid average")
	}

	if a.count, err = strconv.Atoi(r.URL.Query().Get("count")); err != nil {
		return weft.BadRequest("invalid count")
	}

	if a.fifty, err = strconv.Atoi(r.URL.Query().Get("fifty")); err != nil {
		return weft.BadRequest("invalid fifty")
	}

	if a.ninety, err = strconv.Atoi(r.URL.Query().Get("ninety")); err != nil {
		return weft.BadRequest("invalid ninety")
	}

	if a.t, err = time.Parse(time.RFC3339, r.URL.Query().Get("time")); err != nil {
		return weft.BadRequest("invalid time")
	}

	if res := a.applicationSource.loadPK(r); !res.Ok {
		return res
	}

	if res := a.application.loadPK(r); !res.Ok {
		return res
	}

	// TODO - what to do when sending from multiple instances and primary key violations?
	if _, err = db.Exec(`INSERT INTO app.timer(applicationPK, sourcePK, time, average, count, fifty, ninety) VALUES($1,$2,$3,$4,$5,$6,$7)`,
		a.applicationPK, a.sourcePK, a.t, a.average, a.count, a.fifty, a.ninety); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
