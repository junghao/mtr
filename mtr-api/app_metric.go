package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/mtr/internal"
	"github.com/GeoNet/mtr/ts"
	"github.com/GeoNet/weft"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type appMetric struct {
	applicationID string
	applicationPK int
}

type InstanceMetric struct {
	instancePK, typePK int
}

type InstanceMetrics []InstanceMetric

func (l InstanceMetrics) Len() int           { return len(l) }
func (l InstanceMetrics) Less(i, j int) bool { return l[i].instancePK < l[j].instancePK }
func (l InstanceMetrics) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

func (a *appMetric) loadPK(r *http.Request) (res *weft.Result) {
	a.applicationID = r.URL.Query().Get("applicationID")

	err := dbR.QueryRow(`SELECT applicationPK FROM app.application WHERE applicationID = $1`, a.applicationID).Scan(&a.applicationPK)
	switch err {
	case nil:
		return &weft.StatusOK
	case sql.ErrNoRows:
		return &weft.NotFound
	default:
		return weft.InternalServerError(err)
	}
}

func (a *appMetric) svg(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var res *weft.Result
	if res = weft.CheckQuery(r, []string{"applicationID", "group"}, []string{"resolution", "yrange"}); !res.Ok {
		return res
	}

	if res = a.loadPK(r); !res.Ok {
		return res
	}

	var p ts.Plot

	resolution := r.URL.Query().Get("resolution")

	switch resolution {
	case "", "minute":
		resolution = "minute"
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-12), time.Now().UTC())
		p.SetXLabel("12 hours")
	case "five_minutes":
		resolution = "five_minutes"
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*3), time.Now().UTC())
		p.SetXLabel("48 hours")
	case "hour":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*28), time.Now().UTC())
		p.SetXLabel("4 weeks")
	default:
		return weft.BadRequest("invalid value for resolution")
	}

	var err error

	if r.URL.Query().Get("yrange") != "" {
		y := strings.Split(r.URL.Query().Get("yrange"), `,`)

		var ymin, ymax float64

		if len(y) != 2 {
			return weft.BadRequest("invalid yrange query param.")
		}
		if ymin, err = strconv.ParseFloat(y[0], 64); err != nil {
			return weft.BadRequest("invalid yrange query param.")
		}
		if ymax, err = strconv.ParseFloat(y[1], 64); err != nil {
			return weft.BadRequest("invalid yrange query param.")
		}
		p.SetYAxis(ymin, ymax)
	}

	resTitle := resolution
	resTitle = strings.Replace(resTitle, "_", " ", -1)
	resTitle = strings.Title(resTitle)

	switch r.URL.Query().Get("group") {
	case "counters":
		if res := a.loadCounters(resolution, &p); !res.Ok {
			return res
		}

		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Counters - Sum per %s", a.applicationID, resTitle))
		err = ts.MixedAppMetrics.Draw(p, b)
	case "timers":
		if res := a.loadTimers(resolution, &p); !res.Ok {
			return res
		}

		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Timers - 90th Percentile (ms) - Max per %s",
			a.applicationID, resTitle))
		err = ts.MixedAppMetrics.Draw(p, b)
	case "memory":
		if res := a.loadMemory(resolution, &p); !res.Ok {
			return res
		}

		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Memory (bytes) - Average per %s",
			a.applicationID, resTitle))
		err = ts.LineAppMetrics.Draw(p, b)
	case "objects":
		if res := a.loadAppMetrics(resolution, internal.MemHeapObjects, &p); !res.Ok {
			return res
		}

		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Memory Heap Objects (n) - Average per %s",
			a.applicationID, resTitle))
		err = ts.LineAppMetrics.Draw(p, b)
	case "routines":
		if res := a.loadAppMetrics(resolution, internal.Routines, &p); !res.Ok {
			return res
		}
		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Routines (n) - Average per %s",
			a.applicationID, resTitle))
		err = ts.LineAppMetrics.Draw(p, b)
	default:
		return weft.BadRequest("invalid value for type")
	}

	if err != nil {
		return weft.InternalServerError(err)
	}

	h.Set("Content-Type", "image/svg+xml")

	return &weft.StatusOK

}

func (a *appMetric) loadCounters(resolution string, p *ts.Plot) *weft.Result {
	var err error
	var rows *sql.Rows

	switch resolution {
	case "minute":
		rows, err = dbR.Query(`SELECT typePK, date_trunc('`+resolution+`',time) as t, sum(count)
		FROM app.counter WHERE
		applicationPK = $1
		AND time > now() - interval '12 hours'
		GROUP BY date_trunc('`+resolution+`',time), typePK
		ORDER BY t ASC`, a.applicationPK)
	case "five_minutes":
		rows, err = dbR.Query(`SELECT typePK,
		date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t, sum(count)
		FROM app.counter WHERE
		applicationPK = $1
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min', typePK
		ORDER BY t ASC`, a.applicationPK)
	case "hour":
		rows, err = dbR.Query(`SELECT typePK, date_trunc('`+resolution+`',time) as t, sum(count)
		FROM app.counter WHERE
		applicationPK = $1
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time), typePK
		ORDER BY t ASC`, a.applicationPK)
	default:
		return weft.InternalServerError(fmt.Errorf("invalid resolution: %s", resolution))
	}
	if err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var typePK, count int
	pts := make(map[int][]ts.Point)
	total := make(map[int]int)

	for rows.Next() {
		if err = rows.Scan(&typePK, &t, &count); err != nil {
			return weft.InternalServerError(err)
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

	return &weft.StatusOK

}

func (a *appMetric) loadTimers(resolution string, p *ts.Plot) *weft.Result {
	var err error

	var rows *sql.Rows

	switch resolution {
	case "minute":
		rows, err = dbR.Query(`SELECT sourcePK, date_trunc('`+resolution+`',time) as t, max(ninety), sum(count)
		FROM app.timer WHERE
		applicationPK = $1
		AND time > now() - interval '12 hours'
		GROUP BY date_trunc('`+resolution+`',time), sourcePK
		ORDER BY t ASC`, a.applicationPK)
	case "five_minutes":
		rows, err = dbR.Query(`SELECT sourcePK,
		date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t,
		max(ninety), sum(count)
		FROM app.timer WHERE
		applicationPK = $1
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min', sourcePK
		ORDER BY t ASC`, a.applicationPK)
	case "hour":
		rows, err = dbR.Query(`SELECT sourcePK, date_trunc('`+resolution+`',time) as t, max(ninety), sum(count)
		FROM app.timer WHERE
		applicationPK = $1
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time), sourcePK
		ORDER BY t ASC`, a.applicationPK)
	default:
		return weft.InternalServerError(fmt.Errorf("invalid resolution: %s", resolution))
	}
	if err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var sourcePK, avg, n int
	var sourceID string
	pts := make(map[int][]ts.Point)
	total := make(map[int]int)

	for rows.Next() {
		if err = rows.Scan(&sourcePK, &t, &avg, &n); err != nil {
			return weft.InternalServerError(err)
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
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&sourcePK, &sourceID); err != nil {
			return weft.InternalServerError(err)
		}
		sourceIDs[sourcePK] = sourceID
	}
	rows.Close()

	sort.Ints(keys)

	var lables ts.Lables

	for _, k := range keys {
		c := svgColour(sourceIDs[k], k)
		p.AddSeries(ts.Series{Colour: c, Points: pts[k]})
		lables = append(lables, ts.Lable{Colour: c, Lable: fmt.Sprintf("%s (n=%d)", strings.TrimPrefix(sourceIDs[k], `main.`), total[k])})
	}

	p.SetLables(lables)

	return &weft.StatusOK

}

func (a *appMetric) loadMemory(resolution string, p *ts.Plot) *weft.Result {
	var err error

	var rows *sql.Rows

	switch resolution {
	case "minute":
		rows, err = dbR.Query(`SELECT instancePK, typePK, date_trunc('`+resolution+`',time) as t, avg(value)
		FROM app.metric WHERE
		applicationPK = $1 AND typePK IN (1000, 1001, 1002)
		AND time > now() - interval '12 hours'
		GROUP BY date_trunc('`+resolution+`',time), typePK, instancePK
		ORDER BY t ASC`, a.applicationPK)
	case "five_minutes":
		rows, err = dbR.Query(`SELECT instancePK, typePK,
		date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t, avg(value)
		FROM app.metric WHERE
		applicationPK = $1 AND typePK IN (1000, 1001, 1002)
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min', typePK, instancePK
		ORDER BY t ASC`, a.applicationPK)
	case "hour":
		rows, err = dbR.Query(`SELECT instancePK, typePK, date_trunc('`+resolution+`',time) as t, avg(value)
		FROM app.metric WHERE
		applicationPK = $1 AND typePK IN (1000, 1001, 1002)
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time), typePK, instancePK
		ORDER BY t ASC`, a.applicationPK)
	default:
		return weft.InternalServerError(fmt.Errorf("invalid resolution: %s", resolution))
	}
	if err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var typePK, instancePK int
	var avg float64
	var instanceID string
	pts := make(map[InstanceMetric][]ts.Point)

	for rows.Next() {
		if err = rows.Scan(&instancePK, &typePK, &t, &avg); err != nil {
			return weft.InternalServerError(err)
		}
		key := InstanceMetric{instancePK: instancePK, typePK: typePK}
		pts[key] = append(pts[key], ts.Point{DateTime: t, Value: avg})
	}
	rows.Close()

	instanceIDs := make(map[int]string)

	if rows, err = dbR.Query(`SELECT instancePK, instanceID FROM app.instance`); err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&instancePK, &instanceID); err != nil {
			return weft.InternalServerError(err)
		}
		instanceIDs[instancePK] = instanceID
	}
	rows.Close()

	var lables ts.Lables

	for k := range pts {
		p.AddSeries(ts.Series{Colour: internal.Colour(k.typePK), Points: pts[k]})
		lables = append(lables, ts.Lable{Colour: internal.Colour(k.typePK), Lable: fmt.Sprintf("%s.%s", instanceIDs[k.instancePK], strings.TrimPrefix(internal.Lable(k.typePK), `Mem `))})
	}

	p.SetLables(lables)

	return &weft.StatusOK

}

func (a *appMetric) loadAppMetrics(resolution string, typeID internal.ID, p *ts.Plot) *weft.Result {
	var err error

	var rows *sql.Rows

	switch resolution {
	case "minute":
		rows, err = dbR.Query(`SELECT instancePK, typePK, date_trunc('`+resolution+`',time) as t, avg(value)
		FROM app.metric WHERE
		applicationPK = $1 AND typePK = $2
		AND time > now() - interval '12 hours'
		GROUP BY date_trunc('`+resolution+`',time), typePK, instancePK
		ORDER BY t ASC`, a.applicationPK, int(typeID))
	case "five_minutes":
		rows, err = dbR.Query(`SELECT instancePK, typePK,
		date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t, avg(value)
		FROM app.metric WHERE
		applicationPK = $1 AND typePK = $2
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min', typePK, instancePK
		ORDER BY t ASC`, a.applicationPK, int(typeID))
	case "hour":
		rows, err = dbR.Query(`SELECT instancePK, typePK, date_trunc('`+resolution+`',time) as t, avg(value)
		FROM app.metric WHERE
		applicationPK = $1 AND typePK = $2
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time), typePK, instancePK
		ORDER BY t ASC`, a.applicationPK, int(typeID))
	default:
		return weft.InternalServerError(fmt.Errorf("invalid resolution: %s", resolution))
	}
	if err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var typePK, instancePK int
	var avg float64
	var instanceID string
	pts := make(map[InstanceMetric][]ts.Point)

	for rows.Next() {
		if err = rows.Scan(&instancePK, &typePK, &t, &avg); err != nil {
			return weft.InternalServerError(err)
		}
		key := InstanceMetric{instancePK: instancePK, typePK: typePK}
		pts[key] = append(pts[key], ts.Point{DateTime: t, Value: avg})
	}
	rows.Close()

	instanceIDs := make(map[int]string)

	if rows, err = dbR.Query(`SELECT instancePK, instanceID FROM app.instance`); err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&instancePK, &instanceID); err != nil {
			return weft.InternalServerError(err)
		}
		instanceIDs[instancePK] = instanceID
	}
	rows.Close()

	var keys InstanceMetrics

	for k := range pts {
		keys = append(keys, k)
	}

	sort.Sort(keys)

	var lables ts.Lables

	for _, k := range keys {
		c := svgColour(instanceIDs[k.instancePK], k.instancePK)
		p.AddSeries(ts.Series{Colour: c, Points: pts[k]})
		lables = append(lables, ts.Lable{Colour: c, Lable: fmt.Sprintf("%s.%s", instanceIDs[k.instancePK], internal.Lable(k.typePK))})
	}

	p.SetLables(lables)

	return &weft.StatusOK

}

/*
merge merges the output of cs into the single returned chan and waits for all
cs to return.

https://blog.golang.org/pipelines
*/
func merge(cs ...<-chan *weft.Result) <-chan *weft.Result {
	var wg sync.WaitGroup
	out := make(chan *weft.Result)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan *weft.Result) {
		for err := range c {
			out <- err
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
