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

func (a *applicationMetric) save(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"applicationID", "instanceID", "typeID", "time", "value"}, []string{}); !res.Ok {
		return res
	}

	var err error

	// TODO replace other strconv with this approach.
	if a.value, err = strconv.ParseInt(r.URL.Query().Get("value"), 10, 64); err != nil {
		return weft.BadRequest("invalid value")
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

	if res := a.applicationInstance.loadPK(r); !res.Ok {
		return res
	}

	if _, err = db.Exec(`INSERT INTO app.metric (applicationPK, instancePK, typePK, time, value) VALUES($1,$2,$3,$4,$5)`,
		a.applicationPK, a.instancePK, a.typePK, a.t, a.value); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
