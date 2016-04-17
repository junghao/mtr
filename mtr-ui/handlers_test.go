package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func jsonTestServer(jsonInput []byte) (server *httptest.Server) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%s", jsonInput)
	}))

	return testServer
}

func jsonTestServerTearDown(server *httptest.Server) {
	server.Close()
}

// test the function that returns the json []byte's for the given URL
func TestGetBytes(t *testing.T) {
	jsonTestOutput := []byte(`[{"Tag":"1234"}, {"Tag":"ABCD"}]`)
	testServer := jsonTestServer(jsonTestOutput)
	defer jsonTestServerTearDown(testServer)

	jsonObserved, err := getBytes(testServer.URL, "application/json;version=1")
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(jsonTestOutput, jsonObserved) != 0 {
		t.Errorf("expected output: %s, observed %s\n", jsonTestOutput, jsonObserved)
	}
}

func TestGetAllTagIDs(t *testing.T) {
	jsonTestOutput := []byte(`[{"Tag":"1234"}, {"Tag":"ABCD"}]`)
	testServer := jsonTestServer(jsonTestOutput)
	defer jsonTestServerTearDown(testServer)

	allTags, err := getAllTagIDs(testServer.URL)
	if err != nil {
		t.Error(err)
	}

	if allTags[0] != "1234" && allTags[1] != "ABCD" {
		t.Errorf("expected output tags not found\n")
	}
}
