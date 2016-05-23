package main

import (
	"github.com/GeoNet/weft"
	"net/http"
	"strconv"
	"time"
)

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

	if res := a.applicationType.read(v.Get("typeID")); !res.Ok {
		return res
	}

	if res := a.application.readCreate(v.Get("applicationID")); !res.Ok {
		return res
	}

	if res := a.applicationInstance.readCreate(v.Get("instanceID")); !res.Ok {
		return res
	}

	if _, err := db.Exec(`INSERT INTO app.metric (applicationPK, instancePK, typePK, time, value) VALUES($1,$2,$3,$4,$5)`,
		a.application.pk, a.applicationInstance.pk, a.applicationType.pk, a.t, a.value); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
