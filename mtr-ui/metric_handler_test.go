package main

import (
	"net/url"
	"testing"
)

func TestMetricDetailHandler(t *testing.T) {
	var err error
	var tsUrl *url.URL

	tc := &testContext{}
	tc.setup(t)
	defer tc.tearDown()

	if tsUrl, err = url.Parse(tc.testMtrUiServer.URL); err != nil {
		t.Fatal(err)
	}
	tsUrl.Path = "/field/metric"

	// test without a required param, should fail
	doRequest("GET", "text/html", tsUrl.String(), 400, t)

	// add required params.  These don't need to be valid, they're just used in a template for an <img> served by the mtr-api
	q := tsUrl.Query()
	q.Set("deviceID", "dev1")
	q.Set("typeID", "type1")
	tsUrl.RawQuery = q.Encode()
	doRequest("GET", "text/html", tsUrl.String(), 200, t)
}
