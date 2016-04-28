package main

import (
	"testing"
	"net/url"
	"net/http"
	"fmt"
	"bytes"
)

func TestMetricDetailHandler(t *testing.T) {
	// use a mux so we can have custom endpoints and handlerFunc
	var tsUrl *url.URL
	var err error

	tc := &testContext{}
	tc.setup(t)
	defer tc.tearDown()

	// custom handleFunc which emulates the api for getting all tag names
	tc.testMux.HandleFunc("/field/tag", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json;version=1")
		fmt.Fprintf(w, "[{\"Tag\": \"GOVZ\"}, {\"Tag\": \"GRNG\"}]")
	}))

	if tsUrl, err = url.Parse(tc.testServer.URL); err != nil {
		t.Fatal(err)
	}
	q := tsUrl.Query()
	q.Set("deviceID", "dev1")
	q.Set("typeID", "type1")
	tsUrl.RawQuery = q.Encode()

	testRequest := &http.Request{URL: tsUrl}
	testHeader := http.Header{}
	testBuffer := &bytes.Buffer{}
	res := metricDetailHandler(testRequest, testHeader, testBuffer)

	if res.code != http.StatusOK {
		t.Fatalf("response.code from handler is not StatusOK, msg: %s", res.msg)
	}

	if res.ok != true {
		t.Fatalf("response.ok from handler is not true, msg: %s", res.msg)
	}
}