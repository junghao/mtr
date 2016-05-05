/*
Internal defines the wire format for sending and receiving metrics.
*/
package internal

import (
	"time"
)

type AppMetrics struct {
	ApplicationID string // identity for this app e.g., the executable name.
	InstanceID    string // identity for this instance of the application e.g., the host id.
	Metrics       []Metric
	Timers        []Timer
	Counters      []Counter
}

type Metric struct {
	MetricID ID
	Time     time.Time
	Value    int64
}

type Timer struct {
	TimerID string    // An identifier for the thing being timed
	Time    time.Time // Start of the time window.
	Average int32     // in ms
	Count   int32     // Counts for ID in the window.
	Fifty   int32     // the 50th percentile
	Ninety  int32     // the 90th percentile
}

type Counter struct {
	CounterID ID
	Time      time.Time // Start of the time window.
	Count     int32
}
