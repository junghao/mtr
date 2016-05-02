package main

import (
	"bytes"
	"testing"
)

// Exercise all of the templates, any parse error will be raised (not checked at compile time)
func TestTemplates(t *testing.T) {
	b := bytes.Buffer{}

	var sp searchPage
	if err := tagsTemplate.ExecuteTemplate(&b, "border", sp); err != nil {
		t.Error(err)
	}

	var p page
	if err := borderTemplate.ExecuteTemplate(&b, "border", p); err != nil {
		t.Error(err)
	}

	var md metricDetailPage
	if err := metricDetailTemplate.ExecuteTemplate(&b, "border", md); err != nil {
		t.Error(err)
	}
}
