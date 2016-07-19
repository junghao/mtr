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

	expectedVals := [][]string{}
	expectedVals = append(expectedVals, []string{""}) // header line, not checked

	// Load some latency metrics, don't need many
	now := time.Now().UTC()
	v := 140
	for i := -100; i < 0; i += 1 {
		v += i

		ptTime := now.Add(time.Duration(i) * time.Minute)
		r.URL = fmt.Sprintf("/data/latency?siteID=TAUP&typeID=latency.strong&time=%s&mean=%d&fifty=%d&ninety=%d",
			ptTime.Format(time.RFC3339), v, v*10, v*11)

		// expected values
		record := []string{ptTime.Format(DYGRAPH_TIME_FORMAT),
			fmt.Sprintf("%.2f", float32(v)),
			fmt.Sprintf("%.2f", float32(v*10)),
			fmt.Sprintf("%.2f", float32(v*11)),
		}
		expectedVals = append(expectedVals, record)

		if _, err = r.Do(testServer.URL); err != nil {
			t.Error(err)
		}
	}

	r = wt.Request{ID: wt.L(), URL: "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=full", Method: "GET", Accept: "text/csv"}

	var b []byte
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	compareCsvData(b, expectedVals, t)

	// test for no data condition
	r = wt.Request{ID: wt.L(), URL: "/data/latency?siteID=NOT_THERE&typeID=latency.strong&resolution=full", Method: "GET", Accept: "text/csv", Status: http.StatusNotFound}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}
}
