/*
mtrapp for gathering application metrics.
*/
package mtrapp

import (
	"github.com/GeoNet/mtr/internal"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	appName    string
	instanceID string
	once       sync.Once
)

/*
Init initalizes the collection and sending of metrics once per minute if the environment var
MTR_SERVER MTR_USER and MTR_KEY are all non zero the first time it is called.
ApplicationID and InstanceID default to the executable and host names.  These can be set with
the environment var MTR_APPLICATIONID and MTR_INSTANCEID.
*/
func Init() {
	once.Do(func() { collect() })
}

func collect() {
	appName = os.Getenv("MTR_APPLICATIONID")
	instanceID = os.Getenv("MTR_INSTANCEID")

	if appName == "" {
		s := os.Args[0]
		appName = s[strings.LastIndex(s, "/")+1:]
	}

	if instanceID == "" {
		// Not yet clear that this will be enough in a container.
		// Could use an env var or a uuid (would need /dev/random for this)
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
			case <-ticker:
				now = time.Now().UTC()

				runtime.ReadMemStats(&mem)

				m := internal.AppMetrics{
					ApplicationID: appName,
					InstanceID:    instanceID,
					Metrics: []internal.Metric{
						internal.Metric{MetricID: internal.MemSys, Time: now, Value: int64(mem.Sys)},
						internal.Metric{MetricID: internal.MemHeapAlloc, Time: now, Value: int64(mem.HeapAlloc)},
						internal.Metric{MetricID: internal.MemHeapSys, Time: now, Value: int64(mem.HeapSys)},
						internal.Metric{MetricID: internal.MemHeapObjects, Time: now, Value: int64(mem.HeapObjects)},
						internal.Metric{MetricID: internal.Routines, Time: now, Value: int64(runtime.NumGoroutine())},
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
					if v > 0 {
						m.Timers = append(m.Timers, internal.Timer{
							TimerID: k,
							Time:    last,
							Count:   int32(v),
							Total:   int32(sum[k]),
						})
					}
					count[k] = 0
					sum[k] = 0
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
