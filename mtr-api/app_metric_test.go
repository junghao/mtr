package main

import (
	"encoding/csv"
	"fmt"
	wt "github.com/GeoNet/weft/wefttest"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestAppMetricCsv(t *testing.T) {
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

	// Add some app metrics, don't need many
	typeIDs := []int{http.StatusOK,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusNotFound,
		http.StatusOK,
		http.StatusOK,
		http.StatusNotFound,
		http.StatusBadRequest,
		http.StatusInternalServerError,
		http.StatusServiceUnavailable,
	}
	expectedCols := []string{"200", "400", "401", "404", "500", "503"} // the unique numeric equiv of the column names
	now := time.Now().UTC()

	writeTimes := make(map[string]bool)

	expectedVals := make(map[string]map[string]string) // expected results {time:{typeID:value}}
	for i := 0; i < 10; i++ {

		ptTime := now.Add(time.Duration(i) * time.Second)
		r.URL = fmt.Sprintf("/application/counter?applicationID=test-app&instanceID=test-instance&typeID=%d&count=%d&time=%s",
			typeIDs[i], i+1, ptTime.Format(time.RFC3339))

		if _, err = r.Do(testServer.URL); err != nil {
			t.Error(err)
		}

		expectedVals[ptTime.Format(DYGRAPH_TIME_FORMAT)] = map[string]string{strconv.Itoa(typeIDs[i]): fmt.Sprintf("%.2f", float64(i+1))}
		writeTimes[ptTime.Format(DYGRAPH_TIME_FORMAT)] = true
	}

	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=counters", Method: "GET", Accept: "text/csv"}

	var b []byte
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	// for all lines past 0 parse and check values.
	c := csv.NewReader(strings.NewReader(string(b)))
	records, err := c.ReadAll()
	if err == io.EOF {
		t.Error(err)
	}

	for _, record := range records {
		// skip any headers or values that weren't written in this test
		writeTime := record[0]
		if !writeTimes[writeTime] {
			continue
		}

		for f, field := range record {
			if f > 0 && field != "" {
				typeID := expectedCols[f-1]
				if field != expectedVals[writeTime][typeID] {
					t.Errorf("expected %s but got %s for field number %d",
						expectedVals[writeTime][typeID], field, f)
				}
			}
		}
	}

}
