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
	var res *result

	if b, err = ioutil.ReadAll(r.Body); err != nil {
		return internalServerError(err)
	}

	if err = json.Unmarshal(b, &m); err != nil {
		return internalServerError(err)
	}

	var appPK, insPK int

	if appPK, res = applicationPK(m.ApplicationID); !res.ok {
		return res
	}

	if insPK, res = instancePK(m.InstanceID); !res.ok {
		return res
	}

	for _, v := range m.Metrics {
		for i, _ := range resolution {
			if _, err = db.Exec(`INSERT INTO app.metric_`+resolution[i]+`(applicationPK, instancePK, typePK, time, avg, n) VALUES($1,$2,$3,$4,$5,$6)`,
				appPK, insPK, v.MetricID, v.Time.Truncate(duration[i]), v.Value, 1); err != nil {
				if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == `23505` {
					// unique error (already a value at this resolution) update the moving average.
					if _, err = db.Exec(`UPDATE app.metric_`+resolution[i]+` SET avg = ($5 + (avg * n)) / (n+1), n = n + 1
					WHERE applicationPK = $1
					AND instancePK = $2
					AND typePK = $3
					AND time = $4`,
						appPK, insPK, v.MetricID, v.Time.Truncate(duration[i]), v.Value); err != nil {
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
				appPK, v.CounterID, v.Time.Truncate(duration[i]), v.Count); err != nil {
				if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == `23505` {
					// unique error (already a value at this resolution) update the moving average.
					if _, err = db.Exec(`UPDATE app.counter_`+resolution[i]+` SET count = count + $4
					WHERE applicationPK = $1
					AND typePK = $2
					AND time = $3`,
						appPK, v.CounterID, v.Time.Truncate(duration[i]), v.Count); err != nil {
						return internalServerError(err)
					}
				} else {
					return internalServerError(err)
				}
			}
		}
	}

	for _, v := range m.Timers {
		var srcPK int
		if srcPK, res = sourcePK(v.TimerID); !res.ok {
			return res
		}
		for i, _ := range resolution {
			if _, err = db.Exec(`INSERT INTO app.timer_`+resolution[i]+`(applicationPK, sourcePK, time, avg, n) VALUES($1,$2,$3,$4,$5)`,
				appPK, srcPK, v.Time.Truncate(duration[i]), v.Total/v.Count, v.Count); err != nil {
				if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == `23505` {
					// unique error (already a value at this resolution) update the moving average.
					if _, err = db.Exec(`UPDATE app.timer_`+resolution[i]+` SET avg = ($4 + (avg * n)) / (n+$5), n = n + $5
					WHERE applicationPK = $1
					AND sourcePK = $2
					AND time = $3`,
						appPK, srcPK, v.Time.Truncate(duration[i]), v.Total, v.Count); err != nil {
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

// TODO possibly inline these in the save or make it more obvious that they create things as well as looking them up.
// Should not be used from a GET

func applicationPK(applicationID string) (int, *result) {
	var err error
	var pk int

	err = db.QueryRow(`SELECT applicationPK FROM app.application WHERE applicationID = $1`, applicationID).Scan(&pk)

	switch err {
	case nil:
		return pk, &statusOK
	case sql.ErrNoRows:
		if _, err = db.Exec(`INSERT INTO app.application(applicationID) VALUES($1)`, applicationID); err != nil {
			return pk, internalServerError(err)
		}
		if err = db.QueryRow(`SELECT applicationPK FROM app.application WHERE applicationID = $1`, applicationID).Scan(&pk); err != nil {
			return pk, internalServerError(err)
		}
		return pk, &statusOK
	default:
		return pk, internalServerError(err)
	}
}

func instancePK(instanceID string) (int, *result) {
	var err error
	var pk int

	err = db.QueryRow(`SELECT instancePK FROM app.instance WHERE instanceID = $1`, instanceID).Scan(&pk)

	switch err {
	case nil:
		return pk, &statusOK
	case sql.ErrNoRows:
		if _, err = db.Exec(`INSERT INTO app.instance(instanceID) VALUES($1)`, instanceID); err != nil {
			return pk, internalServerError(err)
		}
		if err = db.QueryRow(`SELECT instancePK FROM app.instance WHERE instanceID = $1`, instanceID).Scan(&pk); err != nil {
			return pk, internalServerError(err)
		}
		return pk, &statusOK
	default:
		return pk, internalServerError(err)
	}
}

func sourcePK(sourceID string) (int, *result) {
	var err error
	var pk int

	err = db.QueryRow(`SELECT sourcePK FROM app.source WHERE sourceID = $1`, sourceID).Scan(&pk)

	switch err {
	case nil:
		return pk, &statusOK
	case sql.ErrNoRows:
		if _, err = db.Exec(`INSERT INTO app.source(sourceID) VALUES($1)`, sourceID); err != nil {
			return pk, internalServerError(err)
		}
		if err = db.QueryRow(`SELECT sourcePK FROM app.source WHERE sourceID = $1`, sourceID).Scan(&pk); err != nil {
			return pk, internalServerError(err)
		}
		return pk, &statusOK
	default:
		return pk, internalServerError(err)
	}
}
