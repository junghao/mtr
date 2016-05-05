/*
mtrapp for gathering application metrics.

init initalizes the collection and sending of metrics once per minute if the environment var
MTR_SERVER MTR_USER and MTR_KEY are all non zero.
ApplicationID and InstanceID default to the executable and host names.  These can be set with
the environment var MTR_APPLICATIONID and MTR_INSTANCEID.

Import for side effects  to collect memory and runtime metrics only.
*/
package mtrapp

import (
	"github.com/GeoNet/mtr/internal"
	"log"
	"math"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	appName    string
	instanceID string
	once       sync.Once
)

func init() {
	appName = os.Getenv("MTR_APPLICATIONID")
	instanceID = os.Getenv("MTR_INSTANCEID")

	if appName == "" {
		s := os.Args[0]
		appName = s[strings.LastIndex(s, "/")+1:]
	}

	if instanceID == "" {
		var err error
		instanceID, err = os.Hostname()
		if err != nil {
			log.Println("error finding hostname " + err.Error())
		}
	}

	switch "" {
	case os.Getenv("MTR_SERVER"), os.Getenv("MTR_USER"), os.Getenv("MTR_KEY"):
		log.Println("no mtr credentials, metrics will be dropped.")
	default:
		send = make(chan internal.AppMetrics, 30)
	}

	go sendMetrics(send)

	go func() {
		var mem runtime.MemStats

		ticker := time.NewTicker(time.Minute).C

		var last = time.Now().UTC()
		var now time.Time

		for {
			select {
			case m := <-timers:
				count[m.id]++
				sum[m.id] += m.taken
				taken[m.id] = append(taken[m.id], m.taken)
			case <-ticker:
				now = time.Now().UTC()

				runtime.ReadMemStats(&mem)

				m := internal.AppMetrics{
					ApplicationID: appName,
					InstanceID:    instanceID,
					Metrics: []internal.Metric{
						{MetricID: internal.MemSys, Time: now, Value: int64(mem.Sys)},
						{MetricID: internal.MemHeapAlloc, Time: now, Value: int64(mem.HeapAlloc)},
						{MetricID: internal.MemHeapSys, Time: now, Value: int64(mem.HeapSys)},
						{MetricID: internal.MemHeapObjects, Time: now, Value: int64(mem.HeapObjects)},
						{MetricID: internal.Routines, Time: now, Value: int64(runtime.NumGoroutine())},
					},
				}

				// assume that retrieving values from the counters is fast
				// enough that we don't need a time for each one.
				for i := range counters {
					currVal[i] = counters[i].value()
				}

				for i := range counters {
					if v := currVal[i] - lastVal[i]; v > 0 {
						m.Counters = append(m.Counters, internal.Counter{
							CounterID: counters[i].id,
							Time:      last,
							Count:     int32(v),
						})
					}
				}

				for i := range counters {
					lastVal[i] = currVal[i]
				}

				for k, v := range count {
					m.Timers = append(m.Timers, internal.Timer{
						TimerID: k,
						Time:    last,
						Count:   int32(v),
						Average: int32(sum[k] / v),
						Fifty:   int32(percentile(0.5, taken[k])),
						Ninety:  int32(percentile(0.9, taken[k])),
					})

					delete(taken, k)
					delete(sum, k)
					delete(count, k)
				}

				last = now

				select {
				case send <- m:
				default:
				}
			}
		}
	}()
}

// calculates the kth percentile of v
func percentile(k float64, v []int) (value int) {
	if !sort.IntsAreSorted(v) {
		sort.Ints(v)
	}

	p := k * float64(len(v))

	if p != math.Trunc(p) {
		idx := int(math.Ceil(p))
		if idx <= len(v) {
			value = v[int(math.Ceil(p))-1]
		}
	} else {
		idx := int(math.Trunc(p))
		if idx < len(v) {
			value = int((v[idx-1] + v[idx]) / 2)
		}
	}

	return
}
