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

	utcNow := time.Now().UTC().Truncate(time.Second)
	t0 := utcNow.Add(time.Second * -10)
	testData := []testPoint{
		{time: t0, value: 10000.0}, // can only use one point due to rate limiting in put method
	}

	// the expected CSV data, ignoring the header fields on the first line
	scale := 0.001
	expectedVals := [][]string{
		{""}, // header line, ignored in test.  Should be time, value
		{testData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", testData[0].value*scale)},
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

	// test for invalid deviceID
	r = wt.Request{ID: wt.L(), URL: "/field/metric?deviceID=NOT_THERE&typeID=voltage&resolution=full", Method: "GET", Accept: "text/csv", Status: http.StatusNotFound}
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	// test with a time range, only one point so not much we can do
	start := testData[0].time.Add(time.Second * -1).UTC().Format(time.RFC3339)
	end := testData[0].time.Add(time.Second).UTC().Format(time.RFC3339)
	r = wt.Request{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=full&startDate=" + start + "&endDate=" + end, Method: "GET", Accept: "text/csv"}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	expectedSubset := [][]string{
		{""}, // header line, ignored in test.  Should be time, value
		{testData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", testData[0].value*scale)},
	}
	compareCsvData(b, expectedSubset, t)

	// test multiple typeIDs in a single call (using voltage twice since we only have a single metric), needed for plotting N different series on one graph.
	r = wt.Request{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage&typeID=voltage&resolution=full", Method: "GET", Accept: "text/csv"}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	expectedOutput := [][]string{
		{""}, // header line, ignored in test.  Should be time, voltage, voltage
		{testData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", testData[0].value*scale), fmt.Sprintf("%.2f", testData[0].value*scale)},
	}
	compareCsvData(b, expectedOutput, t)

	// an invalid typeID should get a 404
	r = wt.Request{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=notAValidTypeID&resolution=full", Method: "GET", Accept: "text/csv", Status: http.StatusNotFound}
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

}
