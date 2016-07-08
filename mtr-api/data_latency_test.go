package main

import (
	"encoding/csv"
	"fmt"
	wt "github.com/GeoNet/weft/wefttest"
	"io"
	"strconv"
	"strings"
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

	type latency struct {
		t, v string
	}

	expectedVals := []latency{}

	// Load some latency metrics, don't need many
	now := time.Now().UTC()
	v := 140
	for i := -100; i < 0; i += 1 {
		v += i

		ptTime := now.Add(time.Duration(i) * time.Minute)
		r.URL = fmt.Sprintf("/data/latency?siteID=TAUP&typeID=latency.strong&time=%s&mean=%d",
			ptTime.Format(time.RFC3339), v)

		// expected values
		expectedVals = append(expectedVals, latency{t: ptTime.Format("2006/01/02 15:04:05"), v: strconv.Itoa(v)})

		if _, err = r.Do(testServer.URL); err != nil {
			t.Error(err)
		}
	}

	r = wt.Request{ID: wt.L(), URL: "/data/latency?siteID=TAUP&typeID=latency.strong", Method: "GET", Accept: "text/csv"}

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

	// check csv data, easy to compare all values
	for i, fields := range records {

		// skip headers
		if i == 0 {
			continue
		}

		if fields[0] != expectedVals[i-1].t {
			t.Errorf("expected time=%s but got %s", expectedVals[i-1].t, fields[0])
		}

		if fields[1] != expectedVals[i-1].v {
			t.Errorf("expected value=%s but got %s", expectedVals[i-1].v, fields[1])
		}
	}
}
