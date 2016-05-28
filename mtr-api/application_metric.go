package main

import (
	"github.com/GeoNet/weft"
	"net/http"
	"strconv"
	"time"
)

// appMetric - table app.metric
// things like memory, routines, object count.
type applicationMetric struct {}


// put inserts metrics.  application and instance are added
// to the DB on the fly if required e.g., the first time an
// application sends a metric from an instance.
func (a applicationMetric) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"applicationID", "instanceID", "typeID", "time", "value"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	var err error
	var t     time.Time
	var value int64
	var typePK int

	if value, err = strconv.ParseInt(v.Get("value"), 10, 64); err != nil {
		return weft.BadRequest("invalid value")
	}

	if t, err = time.Parse(time.RFC3339, v.Get("time")); err != nil {
		return weft.BadRequest("invalid time")
	}

	// TODO could validate this from internal
	if typePK, err = strconv.Atoi(v.Get("typeID")); err != nil {
		return weft.BadRequest("invalid typeID")
	}

	applicationID := v.Get("applicationID")
	instanceID := v.Get("instanceID")


	// If we insert one row then return.
	// This will be the most common outcome.
	if result, err := db.Exec(`INSERT INTO app.metric (applicationPK, instancePK, typePK, time, value)
				SELECT applicationPK, instancePK, $3, $4, $5
				FROM app.application, app.instance
				WHERE applicationID = $1
				AND instanceID = $2`,
		applicationID, instanceID, typePK, t, value); err == nil {
		var i int64
		if i, err = result.RowsAffected(); err != nil {
			return weft.InternalServerError(err)
		}
		if i == 1 {
			return &weft.StatusOK
		}
	}

	// Most likely causes of error are missing application or instance.  Add them.
	// Ignore errors - this could race from other handlers.
	db.Exec(`INSERT INTO app.application(applicationID) VALUES($1)`, applicationID)
	db.Exec(`INSERT INTO app.instance(instanceID) VALUES($1)`, instanceID)

	// Try to insert again - if we insert one row then return.
	if result, err := db.Exec(`INSERT INTO app.metric (applicationPK, instancePK, typePK, time, value)
				SELECT applicationPK, instancePK, $3, $4, $5
				FROM app.application, app.instance
				WHERE applicationID = $1
				AND instanceID = $2`,
		applicationID, instanceID, typePK, t, value); err == nil {
		var i int64
		if i, err = result.RowsAffected(); err != nil {
			return weft.InternalServerError(err)
		}
		if i == 1 {
			return &weft.StatusOK
		}
	}

	return weft.InternalServerError(err)
}
