package main

import (
	"github.com/GeoNet/weft"
	"net/http"
	"strconv"
	"time"
	"database/sql"
)

// applicationTimer app.timer
// for timing things.
type applicationTimer struct {}

// put inserts timers.  application and instance are added
// to the DB on the fly if required e.g., the first time an
// application sends a timer from an instance.
func (a applicationTimer) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"applicationID", "instanceID", "sourceID", "time", "average",
		"count", "fifty", "ninety"}, []string{}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	var err error
	var t                             time.Time
	var average, count, fifty, ninety int

	applicationID := v.Get("applicationID")
	instanceID := v.Get("instanceID")
	sourceID := v.Get("sourceID")

	if average, err = strconv.Atoi(v.Get("average")); err != nil {
		return weft.BadRequest("invalid average")
	}

	if count, err = strconv.Atoi(v.Get("count")); err != nil {
		return weft.BadRequest("invalid count")
	}

	if fifty, err = strconv.Atoi(v.Get("fifty")); err != nil {
		return weft.BadRequest("invalid fifty")
	}

	if ninety, err = strconv.Atoi(v.Get("ninety")); err != nil {
		return weft.BadRequest("invalid ninety")
	}

	if t, err = time.Parse(time.RFC3339, v.Get("time")); err != nil {
		return weft.BadRequest("invalid time")
	}

	var result sql.Result

	// If we insert one row then return.
	// This will be the most common outcome.
	if result, err = db.Exec(`INSERT INTO app.timer(applicationPK, instancePK, sourcePK, time, average, count, fifty, ninety)
	 			SELECT applicationPK, instancePK, sourcePK, $4, $5, $6, $7, $8
	 			FROM app.application, app.instance, app.source
				WHERE applicationID = $1
				AND instanceID = $2
				AND sourceID = $3`,
		applicationID, instanceID, sourceID, t, average, count, fifty, ninety); err == nil {
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
	db.Exec(`INSERT INTO app.source(sourceID) VALUES($1)`, sourceID)

	// Try to insert again - if we insert one row then return.
	if result, err = db.Exec(`INSERT INTO app.timer(applicationPK, instancePK, sourcePK, time, average, count, fifty, ninety)
	 			SELECT applicationPK, instancePK, sourcePK, $4, $5, $6, $7, $8
	 			FROM app.application, app.instance, app.source
				WHERE applicationID = $1
				AND instanceID = $2
				AND sourceID = $3`,
		applicationID, instanceID, sourceID, t, average, count, fifty, ninety); err == nil {
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
