package main

import (
//"net/url"
//"reflect"
//"testing"
)

// TODO - remove this?  It's an odd test - it's really testing mtr-api can serve
// data.  This is well tested already.
//func TestGetMatchingMetrics(t *testing.T) {
//	tc := &testContext{}
//	tc.setup(t)
//	defer tc.tearDown()
//
//	u := *mtrApiUrl
//	u.Path = "/tag/GVZ"
//
//	matches, err := getMatchingMetrics(u.String())
//	if err != nil {
//		t.Error(err)
//	}
//
//	expectedType := reflect.TypeOf(metricInfo{})
//
//	if len(matches) <= 0 {
//		t.Errorf("observed metrics length was zero\n")
//	}
//
//	for idx, val := range matches {
//		if reflect.TypeOf(val) != expectedType {
//			t.Errorf("observed metric did not match expected type index:%d, observed val:%s expected val:%s\n", idx, val, expectedType)
//		}
//	}
//}

// TODO - depends on mtr-api state.  Not sure what to do.
//func TestSearchHandler(t *testing.T) {
//	var tsUrl *url.URL
//	var err error
//
//	tc := &testContext{}
//	tc.setup(t)
//	defer tc.tearDown()
//
//	if tsUrl, err = url.Parse(tc.testMtrUiServer.URL); err != nil {
//		t.Fatal(err)
//	}
//	tsUrl.Path = "/search"
//
//	// a request without tagQuery set should fail
//	doRequest("GET", "text/html", tsUrl.String(), 400, t)
//
//	// set tagQuery, should now work
//	q := tsUrl.Query()
//	q.Set("tagQuery", "GVZ")
//	tsUrl.RawQuery = q.Encode()
//	doRequest("GET", "text/html", tsUrl.String(), 200, t)
//}
