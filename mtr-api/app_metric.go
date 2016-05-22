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
	application
}

type InstanceMetric struct {
	instancePK, typePK int
}

type InstanceMetrics []InstanceMetric

func (l InstanceMetrics) Len() int           { return len(l) }
func (l InstanceMetrics) Less(i, j int) bool { return l[i].instancePK < l[j].instancePK }
func (l InstanceMetrics) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

func (a *appMetric) svg(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"applicationID", "group"}, []string{"resolution", "yrange"}); !res.Ok {
		return res
	}

	v := r.URL.Query()

	a.application.id = v.Get("applicationID")

	if res := a.application.read(); !res.Ok {
		return res
	}

	var p ts.Plot

	resolution := v.Get("resolution")

	switch resolution {
	case "", "minute":
		resolution = "minute"
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-12), time.Now().UTC())
		p.SetXLabel("12 hours")
	case "five_minutes":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*3), time.Now().UTC())
		p.SetXLabel("48 hours")
	case "hour":
		p.SetXAxis(time.Now().UTC().Add(time.Hour*-24*28), time.Now().UTC())
		p.SetXLabel("4 weeks")
	default:
		return weft.BadRequest("invalid value for resolution")
	}

	var err error

	if v.Get("yrange") != "" {
		y := strings.Split(v.Get("yrange"), `,`)

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

	switch v.Get("group") {
	case "counters":
		if res := a.loadCounters(resolution, &p); !res.Ok {
			return res
		}

		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Counters - Sum per %s", a.application.id, resTitle))
		err = ts.MixedAppMetrics.Draw(p, b)
	case "timers":
		if res := a.loadTimers(resolution, &p); !res.Ok {
			return res
		}

		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Timers - 90th Percentile (ms) - Max per %s",
			a.application.id, resTitle))
		err = ts.MixedAppMetrics.Draw(p, b)
	case "memory":
		if res := a.loadMemory(resolution, &p); !res.Ok {
			return res
		}

		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Memory (bytes) - Average per %s",
			a.application.id, resTitle))
		err = ts.LineAppMetrics.Draw(p, b)
	case "objects":
		if res := a.loadAppMetrics(resolution, internal.MemHeapObjects, &p); !res.Ok {
			return res
		}

		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Memory Heap Objects (n) - Average per %s",
			a.application.id, resTitle))
		err = ts.LineAppMetrics.Draw(p, b)
	case "routines":
		if res := a.loadAppMetrics(resolution, internal.Routines, &p); !res.Ok {
			return res
		}
		p.SetTitle(fmt.Sprintf("Application: %s, Metric: Routines (n) - Average per %s",
			a.application.id, resTitle))
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
		ORDER BY t ASC`, a.application.pk)
	case "five_minutes":
		rows, err = dbR.Query(`SELECT typePK,
		date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t, sum(count)
		FROM app.counter WHERE
		applicationPK = $1
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min', typePK
		ORDER BY t ASC`, a.application.pk)
	case "hour":
		rows, err = dbR.Query(`SELECT typePK, date_trunc('`+resolution+`',time) as t, sum(count)
		FROM app.counter WHERE
		applicationPK = $1
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time), typePK
		ORDER BY t ASC`, a.application.pk)
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

	var labels ts.Labels

	for _, k := range keys {
		p.AddSeries(ts.Series{Colour: internal.Colour(k), Points: pts[k]})
		labels = append(labels, ts.Label{Colour: internal.Colour(k), Label: fmt.Sprintf("%s (n=%d)", internal.Label(k), total[k])})
	}

	p.SetLabels(labels)

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
		ORDER BY t ASC`, a.application.pk)
	case "five_minutes":
		rows, err = dbR.Query(`SELECT sourcePK,
		date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t,
		max(ninety), sum(count)
		FROM app.timer WHERE
		applicationPK = $1
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min', sourcePK
		ORDER BY t ASC`, a.application.pk)
	case "hour":
		rows, err = dbR.Query(`SELECT sourcePK, date_trunc('`+resolution+`',time) as t, max(ninety), sum(count)
		FROM app.timer WHERE
		applicationPK = $1
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time), sourcePK
		ORDER BY t ASC`, a.application.pk)
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
	rank := make(map[int]int) // track the slowest source value for sorting by colour.

	for rows.Next() {
		if err = rows.Scan(&sourcePK, &t, &avg, &n); err != nil {
			return weft.InternalServerError(err)
		}
		pts[sourcePK] = append(pts[sourcePK], ts.Point{DateTime: t, Value: float64(avg)})
		total[sourcePK] += n
		if rank[sourcePK] < avg {
			rank[sourcePK] = avg
		}
	}
	rows.Close()

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

	// sort the sourcePKs slowest to fastest based on their slowest times.
	keys := rankSlowest(rank)

	// add the time series to the plot. Break the colours up based on source
	// name suffix - GET, PUT, DELETE or other and colour them slowest to fastest.
	// 4 different colours for each group.

	var labels ts.Labels
	var get, put, delete, other int

	for _, k := range keys {
		id := strings.TrimPrefix(sourceIDs[k.Key], `main.`)

		idx := strings.LastIndex(id, ".")
		if idx > -1 && idx+1 < len(id) {
			id = id[idx+1:]
		}

		var c string

		switch id {
		case "GET":
			c = gets[3]
			if get < 3 {
				c = gets[get]
			}
			get++
		case "PUT":
			c = puts[3]
			if put < 3 {
				c = puts[put]
			}
			put++
		case "DELETE":
			c = deletes[3]
			if delete < 3 {
				c = deletes[delete]
			}
			delete++
		default:
			c = colours[3]
			if other < 3 {
				c = colours[other]
			}
			other++
		}

		p.AddSeries(ts.Series{Colour: c, Points: pts[k.Key]})
		labels = append(labels, ts.Label{Colour: c, Label: fmt.Sprintf("%s (n=%d)", strings.TrimPrefix(sourceIDs[k.Key], `main.`), total[k.Key])})
	}

	p.SetLabels(labels)

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
		ORDER BY t ASC`, a.application.pk)
	case "five_minutes":
		rows, err = dbR.Query(`SELECT instancePK, typePK,
		date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t, avg(value)
		FROM app.metric WHERE
		applicationPK = $1 AND typePK IN (1000, 1001, 1002)
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min', typePK, instancePK
		ORDER BY t ASC`, a.application.pk)
	case "hour":
		rows, err = dbR.Query(`SELECT instancePK, typePK, date_trunc('`+resolution+`',time) as t, avg(value)
		FROM app.metric WHERE
		applicationPK = $1 AND typePK IN (1000, 1001, 1002)
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time), typePK, instancePK
		ORDER BY t ASC`, a.application.pk)
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

	var labels ts.Labels

	for k := range pts {
		p.AddSeries(ts.Series{Colour: internal.Colour(k.typePK), Points: pts[k]})
		labels = append(labels, ts.Label{Colour: internal.Colour(k.typePK), Label: fmt.Sprintf("%s.%s", instanceIDs[k.instancePK], strings.TrimPrefix(internal.Label(k.typePK), `Mem `))})
	}

	p.SetLabels(labels)

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
		ORDER BY t ASC`, a.application.pk, int(typeID))
	case "five_minutes":
		rows, err = dbR.Query(`SELECT instancePK, typePK,
		date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min' as t, avg(value)
		FROM app.metric WHERE
		applicationPK = $1 AND typePK = $2
		AND time > now() - interval '2 days'
		GROUP BY date_trunc('hour', time) + extract(minute from time)::int / 5 * interval '5 min', typePK, instancePK
		ORDER BY t ASC`, a.application.pk, int(typeID))
	case "hour":
		rows, err = dbR.Query(`SELECT instancePK, typePK, date_trunc('`+resolution+`',time) as t, avg(value)
		FROM app.metric WHERE
		applicationPK = $1 AND typePK = $2
		AND time > now() - interval '28 days'
		GROUP BY date_trunc('`+resolution+`',time), typePK, instancePK
		ORDER BY t ASC`, a.application.pk, int(typeID))
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

	var labels ts.Labels

	for i, k := range keys {
		if i > len(colours) {
			i = 0
		}

		c := colours[i]

		p.AddSeries(ts.Series{Colour: c, Points: pts[k]})
		labels = append(labels, ts.Label{Colour: c, Label: fmt.Sprintf("%s.%s", instanceIDs[k.instancePK], internal.Label(k.typePK))})
	}

	p.SetLabels(labels)

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

func rankSlowest(r map[int]int) SourceList {
	pl := make(SourceList, len(r))
	i := 0
	for k, v := range r {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   int
	Value int
}

type SourceList []Pair

func (p SourceList) Len() int           { return len(p) }
func (p SourceList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p SourceList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
