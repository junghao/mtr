package main

import (
	"fmt"
	wt "github.com/GeoNet/weft/wefttest"
	"net/http"
	"testing"
	"time"
)

func TestFieldMetricCsv(t *testing.T) {
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

	type testPoint struct {
		time  time.Time
		value float64
	}

	// Testing the "counter" group

	now := time.Now().UTC()
	testData := []testPoint{
		{time: now, value: 1.0}, // can only use one point due to rate limiting in put method
	}

	// the expected CSV data, ignoring the header fields on the first line
	expectedVals := [][]string{
		{""}, // header line, ignored in test.  Should be time, value
		{testData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", testData[0].value)},
	}

	for _, td := range testData {
		r.URL = fmt.Sprintf("/field/metric?deviceID=gps-taupoairport&typeID=voltage&time=%s&value=%d",
			td.time.Format(time.RFC3339), int(td.value))

		addData(r, t)
	}

	r = wt.Request{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=full", Method: "GET", Accept: "text/csv"}

	var b []byte
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}
	compareCsvData(b, expectedVals, t)

	// no data
	r = wt.Request{ID: wt.L(), URL: "/field/metric?deviceID=NOT_THERE&typeID=voltage&resolution=full", Method: "GET", Accept: "text/csv", Status: http.StatusNotFound}
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}
}
