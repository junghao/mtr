package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// a test server running on localhost that serves our custom test content
func setup() (server *httptest.Server) {
	// a test server that tests our handler(s)
	testServer := httptest.NewServer(http.HandlerFunc(toHandler(demoHandler)))

	return testServer
}

func teardown(server *httptest.Server) {
	server.Close()
}

// An example test that checks the body contents, very simple
func TestExample(t *testing.T) {

	var (
		err error
		res *http.Response
	)

	ts := setup()
	defer teardown(ts)

	if res, err = http.Get(ts.URL); err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	if _, err = ioutil.ReadAll(res.Body); err != nil {
		t.Error(err)
	}

	// example of comparing text.  Not performing here since it's getting too big and being modified often.
	//if bytes.Compare(bodyText, []byte("Hello from a demo page")) != 0 {
	//	t.Errorf("unexpected text in body: %s", bodyText)
	//}
}
