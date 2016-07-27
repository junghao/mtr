package main

import (
	"fmt"
	wt "github.com/GeoNet/weft/wefttest"
	"net/http"
	"testing"
	"time"
)

func TestDataLatencyCsv(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	r := wt.Request{
		User:     userW,
		Password: keyW,
		Method:   "PUT",
	}
	var err error

	type latencyTest struct {
		time                time.Time
		mean, fifty, ninety float32
	}

	utcNow := time.Now().UTC().Truncate(time.Second)
	t0 := utcNow.Add(time.Second * -10)
	latencyTestData := []latencyTest{
		{time: t0, mean: 20, fifty: 30, ninety: 40},
		// Can only have one value due to rate_limit.
		// TODO: make the rate_limit value configurable so we can test properly
		//{time: t0.Add(time.Second), mean: 21, fifty:31, ninety: 41},
		//{time: t0.Add(time.Second * 2), mean: 22, fifty:32, ninety: 42},
		//{time: t0.Add(time.Second * 3), mean: 23, fifty:33, ninety: 43},
	}

	expectedVals := [][]string{
		{""}, // header line, ignored in test.
		{latencyTestData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", latencyTestData[0].mean), fmt.Sprintf("%.2f", latencyTestData[0].fifty), fmt.Sprintf("%.2f", latencyTestData[0].ninety)},
		//{latencyTestData[1].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", latencyTestData[1].mean), fmt.Sprintf("%.2f", latencyTestData[1].fifty), fmt.Sprintf("%.2f", latencyTestData[1].ninety)},
		//{latencyTestData[2].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", latencyTestData[2].mean), fmt.Sprintf("%.2f", latencyTestData[2].fifty), fmt.Sprintf("%.2f", latencyTestData[2].ninety)},
		//{latencyTestData[3].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", latencyTestData[3].mean), fmt.Sprintf("%.2f", latencyTestData[3].fifty), fmt.Sprintf("%.2f", latencyTestData[3].ninety)},
	}

	// Add metrics
	for _, lt := range latencyTestData {
		r.URL = fmt.Sprintf("/data/latency?siteID=TAUP&typeID=latency.strong&time=%s&mean=%d&fifty=%d&ninety=%d",
			lt.time.Format(time.RFC3339), int(lt.mean), int(lt.fifty), int(lt.ninety))

		addData(r, t)
	}

	r = wt.Request{ID: wt.L(), URL: "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=full", Method: "GET", Accept: "text/csv"}

	var b []byte
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}
	compareCsvData(b, expectedVals, t)

	// test for invalid siteID condition
	r = wt.Request{ID: wt.L(), URL: "/data/latency?siteID=NOT_THERE&typeID=latency.strong&resolution=full", Method: "GET", Accept: "text/csv", Status: http.StatusNotFound}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	// Test with a time range
	start := latencyTestData[0].time.Add(time.Second * -1).UTC().Format(time.RFC3339)
	end := latencyTestData[0].time.Add(time.Second).UTC().Format(time.RFC3339)
	r = wt.Request{ID: wt.L(), URL: "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=full&startDate=" + start + "&endDate=" + end, Method: "GET", Accept: "text/csv"}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	compareCsvData(b, expectedVals, t)
}
