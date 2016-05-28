package main

import (
	"github.com/GeoNet/weft"
	"net/http"
	"strconv"
	"time"
	"database/sql"
)

// applicationCounter - table app.counter
// things like HTTP requests, messages sent etc.
type applicationCounter struct {}

// put inserts counters.  application and instance are added
// to the DB on the fly if required e.g., the first time an
// application sends a counter from an instance.
func (a applicationCounter) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"applicationID", "instanceID", "typeID", "time", "count"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	var err error
	var t time.Time
	var c, typePK int

	if c, err = strconv.Atoi(v.Get("count")); err != nil {
		return weft.BadRequest("invalid count")
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

	var result sql.Result

	// If we insert one row then return.
	// This will be the most common outcome.
	if result, err = db.Exec(`INSERT INTO app.counter(applicationPK, instancePK, typePK, time, count)
				SELECT applicationPK, instancePK, $3, $4, $5
				FROM app.application, app.instance
				WHERE applicationID = $1
				AND instanceID = $2`,
		applicationID, instanceID, typePK, t, c); err == nil {
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
	if result, err = db.Exec(`INSERT INTO app.counter(applicationPK, instancePK, typePK, time, count)
				SELECT applicationPK, instancePK, $3, $4, $5
				FROM app.application, app.instance
				WHERE applicationID = $1
				AND instanceID = $2`,
		applicationID, instanceID, typePK, t, c); err == nil {
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
