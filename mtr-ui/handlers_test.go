package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type testContext struct {
	testServer      *httptest.Server
	testMux         *http.ServeMux
	actualMtrApiUrl *url.URL
}

// start a test server using a test mux, which we can extend with custom handlerFuncs
func (tc *testContext) setup(t *testing.T) {
	var testUrl *url.URL
	var err error

	tc.testMux = http.NewServeMux()
	tc.testServer = httptest.NewServer(tc.testMux)

	if testUrl, err = url.Parse(tc.testServer.URL); err != nil {
		t.Fatal(err)
	}

	// using our test server as the mtrApiUrl but reverting back at the end of each test
	tc.actualMtrApiUrl = mtrApiUrl
	mtrApiUrl = testUrl
}

func (tc *testContext) tearDown() {
	tc.testServer.Close()
	mtrApiUrl = tc.actualMtrApiUrl
}

func TestCheckQuery(t *testing.T) {
	//checkQuery(r *http.Request, required, optional []string) *result
	var res *result

	// test with no required or optional params
	u := &url.URL{Host:"http://example.com"}
	req := &http.Request{URL:u}
	res = checkQuery(req, []string{}, []string{})
	if res.ok != true {
		t.Fatalf("res.ok not true")
	}

	if res.code != http.StatusOK {
		t.Fatalf("expected res.code to be StatusOK, msg: %s", res.msg)
	}

	// test with cachebuster
	u = &url.URL{Host:"http://example.com", Path:"a_path;something_else"}
	req = &http.Request{URL:u}
	res = checkQuery(req, []string{"required_arg"}, []string{})
	if res.ok != false {
		t.Fatalf("res.ok not false, msg: %s", res.msg)
	}
	if res.code != http.StatusBadRequest {
		t.Fatalf("expected res.code to be StatusBadRequest, msg: %s", res.msg)
	}

	// test with a required param that is missing
	u = &url.URL{Host:"http://example.com"}
	req = &http.Request{URL:u}
	res = checkQuery(req, []string{"required_arg"}, []string{})
	if res.ok != false {
		t.Fatalf("res.ok not false, msg: %s", res.msg)
	}
	if res.code != http.StatusBadRequest {
		t.Fatalf("expected res.code to be StatusBadRequest, msg: %s", res.msg)
	}

	// test with a param that doesn't belong
	u = &url.URL{Host:"http://example.com"}
	q := u.Query()
	q.Set("notright", "true")
	u.RawQuery = q.Encode()

	req = &http.Request{URL:u}
	res = checkQuery(req, []string{}, []string{})
	if res.ok != false {
		t.Fatalf("res.ok not false")
	}
	if res.code != http.StatusBadRequest {
		t.Fatalf("expected res.code to be StatusBadRequest, msg: %s", res.msg)
	}

	// test with a supplied param that is neither required nor optional
	u = &url.URL{Host:"http://example.com"}
	q = u.Query()
	q.Set("stillnotright", "true")
	u.RawQuery = q.Encode()

	req = &http.Request{URL:u}
	res = checkQuery(req, []string{"required_arg1", "required_arg2"}, []string{"optional_arg1", "optional_arg2"})
	if res.ok != false {
		t.Fatalf("res.ok not false, msg: %s", res.msg)
	}
	if res.code != http.StatusBadRequest {
		t.Fatalf("expected res.code to be StatusBadRequest, msg: %s", res.msg)
	}

	// test with a valid required and optional param and an invalid param
	u = &url.URL{Host:"http://example.com"}
	q = u.Query()
	q.Set("required_arg1", "true")
	q.Set("optional_arg1", "true")
	q.Set("stillnotright", "true")
	u.RawQuery = q.Encode()

	req = &http.Request{URL:u}
	res = checkQuery(req, []string{"required_arg1"}, []string{"optional_arg1", "optional_arg2"})
	if res.ok != false {
		t.Fatalf("res.ok not false, msg: %s", res.msg)
	}
	if res.code != http.StatusBadRequest {
		t.Fatalf("expected res.code to be StatusBadRequest, msg: %s", res.msg)
	}

	// test with only valid params
	u = &url.URL{Host:"http://example.com"}
	q = u.Query()
	q.Set("required_arg1", "true")
	q.Set("optional_arg1", "true")
	u.RawQuery = q.Encode()

	req = &http.Request{URL:u}
	res = checkQuery(req, []string{"required_arg1"}, []string{"optional_arg1", "optional_arg2"})
	if res.ok != true {
		t.Fatalf("res.ok not true, msg: %s", res.msg)
	}

	if res.code != http.StatusOK {
		t.Fatalf("expected res.code to be StatusOK, msg: %s", res.msg)
	}
}

// test the function that returns the json []byte's for the given URL
func TestGetBytes(t *testing.T) {
	jsonTestOutput := []byte(`[{"Tag":"1234"}, {"Tag":"ABCD"}]`)
	tc := &testContext{}
	tc.setup(t)
	defer tc.tearDown()

	tc.testMux.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json;version=1")
		w.Write(jsonTestOutput)
	}))

	tc.testMux.HandleFunc("/badrequest", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json;version=1")
	}))

	jsonObserved, err := getBytes(tc.testServer.URL, "application/json;version=1")
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(jsonTestOutput, jsonObserved) != 0 {
		t.Errorf("expected output: %s, observed %s\n", jsonTestOutput, jsonObserved)
	}

	// test with bogus URL
	_, err = getBytes("http://127.0.0.1/not_an_endpoint", "application/json;version=1")
	if err == nil {
		t.Errorf("expected error not returned")
	}
}

func TestGetAllTagIDs(t *testing.T) {
	tc := &testContext{}
	tc.setup(t)
	defer tc.tearDown()

	tc.testMux.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json;version=1")
		fmt.Fprintf(w, `[{"Tag":"1234"}, {"Tag":"ABCD"}]`)
	}))

	allTags, err := getAllTagIDs(tc.testServer.URL)
	if err != nil {
		t.Error(err)
	}

	if len(allTags) == 0 || (allTags[0] != "1234" && allTags[1] != "ABCD") {
		t.Errorf("expected output tags not found\n")
	}
}

// Test the handler directly with mocked out mtr-api server.
func TestDemoHandler(t *testing.T) {
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

	// emulate getting the metric info for a specific tag
	tc.testMux.HandleFunc("/field/metric", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json;version=1")
		fmt.Fprintf(w, "[{\"TypeID\":\"voltage\", \"DeviceID\":\"strong-gretavalley\", \"Tag\":\"GVZ\"}]")
	}))

	if tsUrl, err = url.Parse(tc.testServer.URL); err != nil {
		t.Fatal(err)
	}

	testRequest := &http.Request{URL: tsUrl}
	testHeader := http.Header{}
	testBuffer := &bytes.Buffer{}
	res := demoHandler(testRequest, testHeader, testBuffer)

	if res.code != http.StatusOK {
		t.Fatalf("response.code from handler is not StatusOK, msg: %s", res.msg)
	}

	if res.ok != true {
		t.Fatalf("response.ok not true, msg: %s", res.msg)
	}
}
