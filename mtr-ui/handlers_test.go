package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"strconv"
	"testing"
	"reflect"
)

type testContext struct {
	testMtrUiServer  *httptest.Server
	//testMtrApiServer *httptest.Server
	//testMtrApiMux    *http.ServeMux
	//actualMtrApiUrl  *url.URL
}

// start a test server using a test mux, which we can extend with custom handlerFuncs
func (tc *testContext) setup(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	// our test UI server, uses the same mux as the real UI server
	tc.testMtrUiServer = httptest.NewServer(mux)
}

func (tc *testContext) tearDown() {
	tc.testMtrUiServer.Close()
}

func TestCheckQuery(t *testing.T) {
	//checkQuery(r *http.Request, required, optional []string) *result
	var res *result

	// test with no required or optional params
	u := &url.URL{Host: "http://example.com"}
	req := &http.Request{URL: u}
	res = checkQuery(req, []string{}, []string{})
	if res.ok != true {
		t.Fatalf("res.ok not true")
	}

	if res.code != http.StatusOK {
		t.Fatalf("expected res.code to be StatusOK, msg: %s", res.msg)
	}

	// test with cachebuster
	u = &url.URL{Host: "http://example.com", Path: "a_path;something_else"}
	req = &http.Request{URL: u}
	res = checkQuery(req, []string{"required_arg"}, []string{})
	if res.ok != false {
		t.Fatalf("res.ok not false, msg: %s", res.msg)
	}
	if res.code != http.StatusBadRequest {
		t.Fatalf("expected res.code to be StatusBadRequest, msg: %s", res.msg)
	}

	// test with a required param that is missing
	u = &url.URL{Host: "http://example.com"}
	req = &http.Request{URL: u}
	res = checkQuery(req, []string{"required_arg"}, []string{})
	if res.ok != false {
		t.Fatalf("res.ok not false, msg: %s", res.msg)
	}
	if res.code != http.StatusBadRequest {
		t.Fatalf("expected res.code to be StatusBadRequest, msg: %s", res.msg)
	}

	// test with a param that doesn't belong
	u = &url.URL{Host: "http://example.com"}
	q := u.Query()
	q.Set("notright", "true")
	u.RawQuery = q.Encode()

	req = &http.Request{URL: u}
	res = checkQuery(req, []string{}, []string{})
	if res.ok != false {
		t.Fatalf("res.ok not false")
	}
	if res.code != http.StatusBadRequest {
		t.Fatalf("expected res.code to be StatusBadRequest, msg: %s", res.msg)
	}

	// test with a supplied param that is neither required nor optional
	u = &url.URL{Host: "http://example.com"}
	q = u.Query()
	q.Set("stillnotright", "true")
	u.RawQuery = q.Encode()

	req = &http.Request{URL: u}
	res = checkQuery(req, []string{"required_arg1", "required_arg2"}, []string{"optional_arg1", "optional_arg2"})
	if res.ok != false {
		t.Fatalf("res.ok not false, msg: %s", res.msg)
	}
	if res.code != http.StatusBadRequest {
		t.Fatalf("expected res.code to be StatusBadRequest, msg: %s", res.msg)
	}

	// test with a valid required and optional param and an invalid param
	u = &url.URL{Host: "http://example.com"}
	q = u.Query()
	q.Set("required_arg1", "true")
	q.Set("optional_arg1", "true")
	q.Set("stillnotright", "true")
	u.RawQuery = q.Encode()

	req = &http.Request{URL: u}
	res = checkQuery(req, []string{"required_arg1"}, []string{"optional_arg1", "optional_arg2"})
	if res.ok != false {
		t.Fatalf("res.ok not false, msg: %s", res.msg)
	}
	if res.code != http.StatusBadRequest {
		t.Fatalf("expected res.code to be StatusBadRequest, msg: %s", res.msg)
	}

	// test with only valid params
	u = &url.URL{Host: "http://example.com"}
	q = u.Query()
	q.Set("required_arg1", "true")
	q.Set("optional_arg1", "true")
	u.RawQuery = q.Encode()

	req = &http.Request{URL: u}
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
	//jsonTestOutput := []byte(`[{"Tag":"1234"}, {"Tag":"ABCD"}]`)
	tc := &testContext{}
	tc.setup(t)
	defer tc.tearDown()

	// test getting a byte slice from this URL: https://mtr-api.geonet.org.nz/field/metric/tag?tag=GVZ
	u := *mtrApiUrl
	u.Path = "/field/metric/tag"
	q := u.Query()
	q.Set("tag", "GVZ") // GVZ is a tag with only three metrics (well, at this current moment)
	u.RawQuery = q.Encode()

	getBytesOutput, err := getBytes(u.String(), "application/json;version=1")
	if err != nil {
		t.Error(err)
	}

	// since the output can vary we're just making sure it's non-zero.  TestGetAllTagIDs is a more in-depth test.
	if len(getBytesOutput) <= 0 {
		t.Errorf("getBytes returned an empty slice\n")
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

	u := *mtrApiUrl
	u.Path = "/field/tag"

	allTags, err := getAllTagIDs(u.String())
	if err != nil {
		t.Error(err)
	}

	// Getting results from the live server (not a unit test), so check that all results are strings
	if len(allTags) <= 0 {
		t.Errorf("tag slice is empty")
	}

	for _, val := range allTags {
		if reflect.TypeOf(val).Kind() != reflect.String {
			t.Fatalf("unexpected element type: %s", val)
		}
	}
}

// Test the handler directly with mocked out mtr-api server.
func TestDemoHandler(t *testing.T) {
	var err error
	var tsUrl *url.URL

	tc := &testContext{}
	tc.setup(t)
	defer tc.tearDown()

	if tsUrl, err = url.Parse(tc.testMtrUiServer.URL); err != nil {
		t.Fatal(err)
	}
	tsUrl.Path = "/"
	doRequest("GET", "text/html", tsUrl.String(), 200, t)
}

func doRequest(method, accept, urlString string, status int, t *testing.T) {
	var client = http.Client{}
	var request *http.Request
	var response *http.Response
	var err error
	l := loc()

	if request, err = http.NewRequest(method, urlString, nil); err != nil {
		t.Fatal(err)
	}
	request.Header.Add("Accept", accept)

	if response, err = client.Do(request); err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if status != response.StatusCode {
		t.Errorf("Wrong response code for %s got %d expected %d", l, response.StatusCode, status)
	}

	if method == "GET" && status == http.StatusOK {
		if response.Header.Get("Content-Type") != accept {
			t.Errorf("Wrong Content-Type for %s got %s expected %s", l, response.Header.Get("Content-Type"), accept)
		}
	}
}

// loc returns a string representing the line of code 2 functions calls back.
func loc() (loc string) {
	_, _, l, _ := runtime.Caller(2)
	return "L" + strconv.Itoa(l)
}
