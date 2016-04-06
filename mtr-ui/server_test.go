package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// a test server running on localhost that serves our custom test content
func setup() (server *httptest.Server) {
	// a test server that tests our handler(s)
	testServer := httptest.NewServer(http.HandlerFunc(toHandler(handler)))

	return testServer
}

func teardown(server *httptest.Server) {
	server.Close()
}

// An example test that checks the body contents, very simple
func TestExample(t *testing.T) {

	ts := setup()
	defer teardown(ts)

	res, err := http.Get(ts.URL)
	defer res.Body.Close()
	if err != nil {
		t.Error(err)
	}

	bodyText, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(bodyText, []byte("Hello from a demo page")) != 0 {
		t.Errorf("unexpected text in body: %s", bodyText)
	}

}
