package main

import (
	"github.com/GeoNet/weft"
	"net/http"
	"strconv"
	"time"
)

type applicationMetric struct {
	application
	applicationInstance
	applicationType
	t     time.Time
	value int64
}

// create saves the application metric to the db.
func (a *applicationMetric) create() *weft.Result {
	if res := a.applicationType.read(); !res.Ok {
		return res
	}

	if res := a.application.create(); !res.Ok {
		return res
	}

	if res := a.applicationInstance.create(); !res.Ok {
		return res
	}

	if _, err := db.Exec(`INSERT INTO app.metric (applicationPK, instancePK, typePK, time, value) VALUES($1,$2,$3,$4,$5)`,
		a.application.pk, a.applicationInstance.pk, a.applicationType.pk, a.t, a.value); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

// put handles http.PUT methods, parsing the request and saving to the db.
func (a *applicationMetric) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"applicationID", "instanceID", "typeID", "time", "value"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	var err error

	if a.value, err = strconv.ParseInt(v.Get("value"), 10, 64); err != nil {
		return weft.BadRequest("invalid value")
	}

	if a.t, err = time.Parse(time.RFC3339, v.Get("time")); err != nil {
		return weft.BadRequest("invalid time")
	}

	a.applicationType.id = v.Get("typeID")
	a.application.id = v.Get("applicationID")
	a.applicationInstance.id = v.Get("instanceID")

	return a.create()
}
