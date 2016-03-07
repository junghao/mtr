package main

import (
	"database/sql"
	"encoding/json"
	"github.com/GeoNet/mtr/internal"
	"github.com/lib/pq"
	"io/ioutil"
	"net/http"
)

type appMetric struct{}

func (a *appMetric) save(r *http.Request) *result {
	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	var b []byte
	var err error
	var m internal.AppMetrics

	if b, err = ioutil.ReadAll(r.Body); err != nil {
		return internalServerError(err)
	}

	if err = json.Unmarshal(b, &m); err != nil {
		return internalServerError(err)
	}

	// Find  (and possibly create) the applicationPK for the applicationID
	var applicationPK int

	err = db.QueryRow(`SELECT applicationPK FROM app.application WHERE applicationID = $1`, m.ApplicationID).Scan(&applicationPK)
	switch err {
	case nil:
	case sql.ErrNoRows:
		if _, err = db.Exec(`INSERT INTO app.application(applicationID) VALUES($1)`, m.ApplicationID); err != nil {
			return internalServerError(err)
		}
		if err = db.QueryRow(`SELECT applicationPK FROM app.application WHERE applicationID = $1`, m.ApplicationID).Scan(&applicationPK); err != nil {
			return internalServerError(err)
		}
	default:
		return internalServerError(err)
	}

	// Find  (and possibly create) the instancePK for the instanceID
	var instancePK int

	err = db.QueryRow(`SELECT instancePK FROM app.instance WHERE instanceID = $1`, m.InstanceID).Scan(&instancePK)
	switch err {
	case nil:
	case sql.ErrNoRows:
		if _, err = db.Exec(`INSERT INTO app.instance(instanceID) VALUES($1)`, m.InstanceID); err != nil {
			return internalServerError(err)
		}
		if err = db.QueryRow(`SELECT instancePK FROM app.instance WHERE instanceID = $1`, m.InstanceID).Scan(&instancePK); err != nil {
			return internalServerError(err)
		}
	default:
		return internalServerError(err)
	}

	for _, v := range m.Metrics {
		for i, _ := range resolution {
			if _, err = db.Exec(`INSERT INTO app.metric_`+resolution[i]+`(applicationPK, instancePK, typePK, time, avg, n) VALUES($1,$2,$3,$4,$5,$6)`,
				applicationPK, instancePK, v.MetricID, v.Time.Truncate(duration[i]), v.Value, 1); err != nil {
				if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == `23505` {
					// unique error (already a value at this resolution) update the moving average.
					if _, err = db.Exec(`UPDATE app.metric_`+resolution[i]+` SET avg = ($5 + (avg * n)) / (n+1), n = n + 1
					WHERE applicationPK = $1
					AND instancePK = $2
					AND typePK = $3
					AND time = $4`,
						applicationPK, instancePK, v.MetricID, v.Time.Truncate(duration[i]), v.Value); err != nil {
						return internalServerError(err)
					}
				} else {
					return internalServerError(err)
				}
			}
		}
	}

	for _, v := range m.Counters {
		for i, _ := range resolution {
			if _, err = db.Exec(`INSERT INTO app.counter_`+resolution[i]+`(applicationPK, typePK, time, count) VALUES($1,$2,$3,$4)`,
				applicationPK, v.CounterID, v.Time.Truncate(duration[i]), v.Count); err != nil {
				if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == `23505` {
					// unique error (already a value at this resolution) update the moving average.
					if _, err = db.Exec(`UPDATE app.counter_`+resolution[i]+` SET count = count + $4
					WHERE applicationPK = $1
					AND typePK = $2
					AND time = $3`,
						applicationPK, v.CounterID, v.Time.Truncate(duration[i]), v.Count); err != nil {
						return internalServerError(err)
					}
				} else {
					return internalServerError(err)
				}
			}
		}
	}

	for _, v := range m.Timers {
		// Find  (and possibly create) the sourcePK for the sourceID
		var sourcePK int

		err = db.QueryRow(`SELECT sourcePK FROM app.source WHERE sourceID = $1`, v.TimerID).Scan(&sourcePK)

		switch err {
		case nil:
		case sql.ErrNoRows:
			if _, err = db.Exec(`INSERT INTO app.source(sourceID) VALUES($1)`, v.TimerID); err != nil {
				return internalServerError(err)
			}
			if err = db.QueryRow(`SELECT sourcePK FROM app.source WHERE sourceID = $1`, v.TimerID).Scan(&sourcePK); err != nil {
				return internalServerError(err)
			}
		default:
			return internalServerError(err)
		}

		for i, _ := range resolution {
			if _, err = db.Exec(`INSERT INTO app.timer_`+resolution[i]+`(applicationPK, sourcePK, time, avg, n) VALUES($1,$2,$3,$4,$5)`,
				applicationPK, sourcePK, v.Time.Truncate(duration[i]), v.Total/v.Count, v.Count); err != nil {
				if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == `23505` {
					// unique error (already a value at this resolution) update the moving average.
					if _, err = db.Exec(`UPDATE app.timer_`+resolution[i]+` SET avg = ($4 + (avg * n)) / (n+$5), n = n + $5
					WHERE applicationPK = $1
					AND sourcePK = $2
					AND time = $3`,
						applicationPK, sourcePK, v.Time.Truncate(duration[i]), v.Total, v.Count); err != nil {
						return internalServerError(err)
					}
				} else {
					return internalServerError(err)
				}
			}
		}
	}

	return &statusOK
}
