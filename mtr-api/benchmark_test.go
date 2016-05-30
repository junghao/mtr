package main

import (
	"net/http"
	"strconv"
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

	if r := delApplication("test-app-bench"); !r.Ok {
		b.Error(r.Msg)
	}

	var req *http.Request
	var err error
	if req, err = http.NewRequest("PUT", "http://test.com/application/metric", nil); err != nil {
		b.Fatal(err)
	}

	t := time.Now().UTC()

	q := req.URL.Query()
	q.Add("applicationID", "test-app-bench")
	q.Add("instanceID", "test-app-bench-instance")
	q.Add("typeID", strconv.Itoa(1000))
	q.Add("value", strconv.FormatInt(12000, 10))

	var a applicationMetric

	for n := 0; n < b.N; n++ {
		t = t.Add(time.Second)
		q.Set("time", t.Format(time.RFC3339))
		req.URL.RawQuery = q.Encode()

		if res := a.put(req); !res.Ok {
			b.Error(res.Msg)
		}
	}
}

func BenchmarkFieldMetric(b *testing.B) {
	setupBench(b)
	defer teardown()

	// Delete any benchmark data from the DB by deleting
	// the device model and then recreate it for the test.

	var req *http.Request
	var err error

	if req, err = http.NewRequest("PUT", "http://test.com/field/model", nil); err != nil {
		b.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("modelID", "device-model-bench")
	req.URL.RawQuery = q.Encode()

	var dm fieldModel
	if res := dm.delete(req); !res.Ok {
		b.Fatal(res.Msg)
	}

	if res := dm.put(req); !res.Ok {
		b.Fatal(res.Msg)
	}

	if req, err = http.NewRequest("PUT", "http://test.com/field/device", nil); err != nil {
		b.Fatal(err)
	}

	q = req.URL.Query()
	q.Add("deviceID", "device-bench")
	q.Add("modelID", "device-model-bench")
	q.Add("latitude", "12")
	q.Add("longitude", "12")
	req.URL.RawQuery = q.Encode()

	var d fieldDevice
	if res := d.put(req); !res.Ok {
		b.Fatal(res.Msg)
	}

	// Benchmark sending field metrics

	if req, err = http.NewRequest("PUT", "http://test.com/field/metric", nil); err != nil {
		b.Fatal(err)
	}

	q = req.URL.Query()
	q.Add("deviceID", "device-bench")
	q.Add("typeID", "voltage")
	q.Add("value", "14100")

	t := time.Now().UTC()

	var fm fieldMetric

	for n := 0; n < b.N; n++ {
		t = t.Add(time.Minute)
		q.Set("time", t.Format(time.RFC3339))
		req.URL.RawQuery = q.Encode()

		if res := fm.put(req); !res.Ok {
			b.Error(res.Msg)
		}
	}

	// Delete the benchmark data from the DB

	if req, err = http.NewRequest("PUT", "http://test.com/field/model", nil); err != nil {
		b.Fatal(err)
	}

	q = req.URL.Query()
	q.Add("modelID", "device-model-bench")
	req.URL.RawQuery = q.Encode()

	if res := dm.delete(req); !res.Ok {
		b.Fatal(res.Msg)
	}

}
