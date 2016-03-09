package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/GeoNet/mtr/internal"
	"github.com/GeoNet/mtr/ts"
	"github.com/lib/pq"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

var appResolution = [...]string{
	"minute",
	"hour",
}

var appDuration = [...]time.Duration{
	time.Minute,
	time.Hour,
}

var colours = [...]string{
	"deepskyblue",
	"darkcyan",
	"darkgoldenrod",
	"lawngreen",
	"orangered",
	"darkcyan",
	"forestgreen",
	"mediumslateblue",
}

var numColours = len(colours) - 1

type appMetric struct {
	applicationID string
	applicationPK int
}

type instanceMetric struct {
	instancePK, typePK int
}

func (a *appMetric) loadPK(r *http.Request) (res *result) {
	a.applicationID = r.URL.Query().Get("applicationID")

	err := dbR.QueryRow(`SELECT applicationPK FROM app.application WHERE applicationID = $1`, a.applicationID).Scan(&a.applicationPK)
	switch err {
	case nil:
		return &statusOK
	case sql.ErrNoRows:
		return &notFound
	default:
		return internalServerError(err)
	}
}

/*
Handles requests like
/app/metric?applicationID=mtr-api&group=timers
/app/metric?applicationID=mtr-api&group=counters
/app/metric?applicationID=mtr-api&group=memory
/app/metric?applicationID=mtr-api&group=objects
/app/metric?applicationID=mtr-api&group=routines

Metrics are available at minute (default) and hour resolution.
*/
func (a *appMetric) svg(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	var res *result
	if res = checkQuery(r, []string{"applicationID", "group"}, []string{"resolution"}); !res.ok {
		return res
	}

	if res = a.loadPK(r); !res.ok {
		return res
	}

	var p ts.Plot

	resolution := r.URL.Query().Get("resolution")

	switch resolution {
	case "", "minute":
		resolution = "minute"
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-12), time.Now().UTC())
		p.SetXLabel("12 hours")
	case "hour":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*28), time.Now().UTC())
		p.SetXLabel("4 weeks")
	default:
		return badRequest("invalid value for resolution")
	}

	var err error

	switch r.URL.Query().Get("group") {
	case "counters":
		if res := a.loadCounters(resolution, &p); !res.ok {
			return res
		}

		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Counters", a.applicationID))
	case "timers":
		if res := a.loadTimers(resolution, &p); !res.ok {
			return res
		}

		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Timers (ms)", a.applicationID))
	case "memory":
		if res := a.loadMemory(resolution, &p); !res.ok {
			return res
		}

		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Memory (bytes)", a.applicationID))
	case "objects":
		if res := a.loadAppMetrics(resolution, internal.MemHeapObjects, &p); !res.ok {
			return res
		}

		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Memory Heap Objects (n)", a.applicationID))
	case "routines":
		if res := a.loadAppMetrics(resolution, internal.Routines, &p); !res.ok {
			return res
		}
		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Routines (n)", a.applicationID))

	default:
		return badRequest("invalid value for type")
	}

	err = ts.LineAppMetrics.Draw(p, b)

	if err != nil {
		return internalServerError(err)
	}

	h.Set("Content-Type", "image/svg+xml")

	return &statusOK

}

func (a *appMetric) loadCounters(resolution string, p *ts.Plot) *result {
	var err error

	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT typePK, time, count FROM app.counter_`+resolution+` WHERE 
		applicationPK = $1
		ORDER BY time ASC`, a.applicationPK); err != nil {
		return internalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var typePK, count int
	pts := make(map[int][]ts.Point)
	total := make(map[int]int)

	for rows.Next() {
		if err = rows.Scan(&typePK, &t, &count); err != nil {
			return internalServerError(err)
		}
		pts[typePK] = append(pts[typePK], ts.Point{DateTime: t, Value: float64(count)})
		total[typePK] += count
	}
	rows.Close()

	var keys []int
	for k := range pts {
		keys = append(keys, k)

	}

	sort.Ints(keys)

	var lables ts.Lables

	for _, k := range keys {
		p.AddSeries(ts.Series{Colour: internal.Colour(k), Points: pts[k]})
		lables = append(lables, ts.Lable{Colour: internal.Colour(k), Lable: fmt.Sprintf("%s (n=%d)", internal.Lable(k), total[k])})
	}

	p.SetLables(lables)

	return &statusOK

}

func (a *appMetric) loadTimers(resolution string, p *ts.Plot) *result {
	var err error

	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT sourcePK, time, avg, n FROM app.timer_`+resolution+` WHERE 
		applicationPK = $1
		ORDER BY time ASC`, a.applicationPK); err != nil {
		return internalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var sourcePK, avg, n int
	var sourceID string
	pts := make(map[int][]ts.Point)
	total := make(map[int]int)

	for rows.Next() {
		if err = rows.Scan(&sourcePK, &t, &avg, &n); err != nil {
			return internalServerError(err)
		}
		pts[sourcePK] = append(pts[sourcePK], ts.Point{DateTime: t, Value: float64(avg)})
		total[sourcePK] += n
	}
	rows.Close()

	var keys []int
	for k := range pts {
		keys = append(keys, k)

	}

	sourceIDs := make(map[int]string)

	if rows, err = dbR.Query(`SELECT sourcePK, sourceID FROM app.source`); err != nil {
		return internalServerError(err)
	}

	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&sourcePK, &sourceID); err != nil {
			return internalServerError(err)
		}
		sourceIDs[sourcePK] = sourceID
	}
	rows.Close()

	sort.Ints(keys)

	var lables ts.Lables

	for i, k := range keys {
		if i > numColours {
			i = 0
		}
		p.AddSeries(ts.Series{Colour: colours[i], Points: pts[k]})
		lables = append(lables, ts.Lable{Colour: colours[i], Lable: fmt.Sprintf("%s (n=%d)", strings.TrimPrefix(sourceIDs[k], `main.`), total[k])})
	}

	p.SetLables(lables)

	return &statusOK

}

func (a *appMetric) loadMemory(resolution string, p *ts.Plot) *result {
	var err error

	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT instancePK, typePK, time, avg FROM app.metric_`+resolution+` WHERE 
		applicationPK = $1 AND typePK IN (1000, 1001, 1002) 
		ORDER BY time ASC`, a.applicationPK); err != nil {
		return internalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var typePK, instancePK, avg int
	var instanceID string
	pts := make(map[instanceMetric][]ts.Point)

	for rows.Next() {
		if err = rows.Scan(&instancePK, &typePK, &t, &avg); err != nil {
			return internalServerError(err)
		}
		key := instanceMetric{instancePK: instancePK, typePK: typePK}
		pts[key] = append(pts[key], ts.Point{DateTime: t, Value: float64(avg)})
	}
	rows.Close()

	var keys []int
	for k := range pts {
		keys = append(keys, k.instancePK)

	}

	instanceIDs := make(map[int]string)

	if rows, err = dbR.Query(`SELECT instancePK, instanceID FROM app.instance`); err != nil {
		return internalServerError(err)
	}

	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&instancePK, &instanceID); err != nil {
			return internalServerError(err)
		}
		instanceIDs[instancePK] = instanceID
	}
	rows.Close()

	sort.Ints(keys)

	var lables ts.Lables

	for k, _ := range pts {
		p.AddSeries(ts.Series{Colour: internal.Colour(k.typePK), Points: pts[k]})
		lables = append(lables, ts.Lable{Colour: internal.Colour(k.typePK), Lable: fmt.Sprintf("%s.%s", instanceIDs[k.instancePK], strings.TrimPrefix(internal.Lable(k.typePK), `Mem `))})
	}

	p.SetLables(lables)

	return &statusOK

}

func (a *appMetric) loadAppMetrics(resolution string, typeID internal.ID, p *ts.Plot) *result {
	var err error

	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT instancePK, typePK, time, avg FROM app.metric_`+resolution+` WHERE 
		applicationPK = $1 AND typePK = $2 
		ORDER BY time ASC`, a.applicationPK, int(typeID)); err != nil {
		return internalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var typePK, instancePK, avg int
	var instanceID string
	pts := make(map[instanceMetric][]ts.Point)

	for rows.Next() {
		if err = rows.Scan(&instancePK, &typePK, &t, &avg); err != nil {
			return internalServerError(err)
		}
		key := instanceMetric{instancePK: instancePK, typePK: typePK}
		pts[key] = append(pts[key], ts.Point{DateTime: t, Value: float64(avg)})
	}
	rows.Close()

	var keys []int
	for k := range pts {
		keys = append(keys, k.instancePK)

	}

	instanceIDs := make(map[int]string)

	if rows, err = dbR.Query(`SELECT instancePK, instanceID FROM app.instance`); err != nil {
		return internalServerError(err)
	}

	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&instancePK, &instanceID); err != nil {
			return internalServerError(err)
		}
		instanceIDs[instancePK] = instanceID
	}
	rows.Close()

	sort.Ints(keys)

	var lables ts.Lables

	for k, _ := range pts {
		p.AddSeries(ts.Series{Colour: internal.Colour(k.typePK), Points: pts[k]})
		lables = append(lables, ts.Lable{Colour: internal.Colour(k.typePK), Lable: fmt.Sprintf("%s.%s", instanceIDs[k.instancePK], internal.Lable(k.typePK))})
	}

	p.SetLables(lables)

	return &statusOK

}

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
		for i, _ := range appResolution {
			if _, err = db.Exec(`INSERT INTO app.metric_`+appResolution[i]+`(applicationPK, instancePK, typePK, time, avg, n) VALUES($1,$2,$3,$4,$5,$6)`,
				applicationPK, instancePK, v.MetricID, v.Time.Truncate(appDuration[i]), v.Value, 1); err != nil {
				if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == errorUniqueViolation {
					// unique error (already a value at this resolution) update the moving average.
					if _, err = db.Exec(`UPDATE app.metric_`+appResolution[i]+` SET avg = ($5 + (avg * n)) / (n+1), n = n + 1
					WHERE applicationPK = $1
					AND instancePK = $2
					AND typePK = $3
					AND time = $4`,
						applicationPK, instancePK, v.MetricID, v.Time.Truncate(appDuration[i]), v.Value); err != nil {
						return internalServerError(err)
					}
				} else {
					return internalServerError(err)
				}
			}
		}
	}

	for _, v := range m.Counters {
		for i, _ := range appResolution {
			if _, err = db.Exec(`INSERT INTO app.counter_`+appResolution[i]+`(applicationPK, typePK, time, count) VALUES($1,$2,$3,$4)`,
				applicationPK, v.CounterID, v.Time.Truncate(appDuration[i]), v.Count); err != nil {
				if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == errorUniqueViolation {
					// unique error (already a value at this resolution) update the moving average.
					if _, err = db.Exec(`UPDATE app.counter_`+appResolution[i]+` SET count = count + $4
					WHERE applicationPK = $1
					AND typePK = $2
					AND time = $3`,
						applicationPK, v.CounterID, v.Time.Truncate(appDuration[i]), v.Count); err != nil {
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

		for i, _ := range appResolution {
			if _, err = db.Exec(`INSERT INTO app.timer_`+appResolution[i]+`(applicationPK, sourcePK, time, avg, n) VALUES($1,$2,$3,$4,$5)`,
				applicationPK, sourcePK, v.Time.Truncate(appDuration[i]), v.Total/v.Count, v.Count); err != nil {
				if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == errorUniqueViolation {
					// unique error (already a value at this resolution) update the moving average.
					if _, err = db.Exec(`UPDATE app.timer_`+appResolution[i]+` SET avg = ($4 + (avg * n)) / (n+$5), n = n + $5
					WHERE applicationPK = $1
					AND sourcePK = $2
					AND time = $3`,
						applicationPK, sourcePK, v.Time.Truncate(appDuration[i]), v.Total, v.Count); err != nil {
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
