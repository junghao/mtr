package main

import (
	"testing"
	"time"
)

// without memory cache for pk:
// BenchmarkApplicationMetricCreate-4	     500	   3425893 ns/op
//
// with memory cache for pk:
// BenchmarkApplicationMetricCreate-4	     500	   2478556 ns/op
func BenchmarkApplicationMetricCreate(b *testing.B) {
	setupBench(b)
	defer teardown()

	var a applicationMetric

	a.application.id = "test-app-bench"
	a.applicationInstance.id = "test-app-bench-instance"
	a.applicationType.id = "1000"
	a.value = 12000
	a.t = time.Now().UTC()

	if r := a.application.del(); !r.Ok {
		b.Error(r.Msg)
	}

	for n := 0; n < b.N; n++ {
		if r := a.create(); !r.Ok {
			b.Error(r.Msg)
		}
		a.t = a.t.Add(time.Second)
	}
}
