package main

import (
	"net/http"
	"net/url"
	"testing"
)

func TestGetMatchingMetrics(t *testing.T) {
	jsonTestOutput := []byte(`[{"TypeID":"voltage", "DeviceID":"companyA", "Tag":"1234"}, {"TypeID":"voltage", "DeviceID":"companyB", "Tag":"ABCD"}]`)
	tc := &testContext{}
	tc.setup(t)
	defer tc.tearDown()

	tc.testMtrApiMux.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json;version=1")
		w.Write(jsonTestOutput)
	}))

	matches, err := getMatchingMetrics(tc.testMtrApiServer.URL)
	if err != nil {
		t.Error(err)
	}

	expectedMetrics := matchingMetrics{metricInfo{TypeID: "voltage", DeviceID: "companyA", Tag: "1234"},
		metricInfo{TypeID: "voltage", DeviceID: "companyB", Tag: "ABCD"}}

	if len(matches) != len(expectedMetrics) {
		t.Errorf("observed metrics length: %d did not match expected length: %d\n", len(matches), len(expectedMetrics))
	}

	for idx, val := range matches {
		expect := expectedMetrics[idx]
		// compare all struct members apart from []bytes
		if val.DeviceID != expect.DeviceID || val.Tag != expect.Tag || val.TypeID != expect.TypeID {
			t.Errorf("observed metric did not match expected for index %d\n", idx)
		}
	}
}

func TestSearchHandler(t *testing.T) {
	var tsUrl *url.URL
	var err error

	tc := &testContext{}
	tc.setup(t)
	defer tc.tearDown()

	// custom handleFunc which emulates the api for getting all tag names
	tc.testMtrApiMux.HandleFunc("/field/tag", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json;version=1")
		//fmt.Fprintf(w, "[{\"Tag\": \"GOVZ\"}, {\"Tag\": \"GRNG\"}]")
		w.Write([]byte(`[{"Tag": "GOVZ"}, {"Tag": "GRNG"}]`))
	}))

	// serve fake metric data for a given tag
	tc.testMtrApiMux.HandleFunc("/field/metric/tag", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json;version=1")
		//fmt.Fprintf(w, "[{\"Tag\": \"GOVZ\"}, {\"Tag\": \"GRNG\"}]")
		w.Write([]byte(`[{"TypeID":"voltage", "DeviceID":"companyA", "Tag":"TAGX"}, {"TypeID":"voltage", "DeviceID":"companyB", "Tag":"TAGX"}]`))
	}))

	if tsUrl, err = url.Parse(tc.testMtrUiServer.URL); err != nil {
		t.Fatal(err)
	}
	tsUrl.Path = "/search"

	// a request without tagQuery set should fail
	doRequest("GET", "text/html", tsUrl.String(), 400, t)

	// set tagQuery, should now work
	q := tsUrl.Query()
	q.Set("tagQuery", "TAGX")
	tsUrl.RawQuery = q.Encode()
	doRequest("GET", "text/html", tsUrl.String(), 200, t)
}
