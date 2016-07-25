package main

import (
	"bytes"
	"testing"
)

// Exercise all of the templates, any parse error will be raised (not checked at compile time)
func TestTemplates(t *testing.T) {
	b := bytes.Buffer{}

	var sp searchPage
	if err := tagSearchTemplate.ExecuteTemplate(&b, "border", sp); err != nil {
		t.Error(err)
	}

	var p mtrUiPage
	if err := homepageTemplate.ExecuteTemplate(&b, "border", p); err != nil {
		t.Error(err)
	}
	if err := fieldTemplate.ExecuteTemplate(&b, "border", p); err != nil {
		t.Error(err)
	}
	if err := dataTemplate.ExecuteTemplate(&b, "border", p); err != nil {
		t.Error(err)
	}

	var mp mapPage
	if err := mapTemplate.ExecuteTemplate(&b, "border", mp); err != nil {
		t.Error(err)
	}

	var tp tagPage
	if err := tagPageTemplate.ExecuteTemplate(&b, "border", tp); err != nil {
		t.Error(err)
	}

	var md metricDetailPage
	if err := metricDetailTemplate.ExecuteTemplate(&b, "border", md); err != nil {
		t.Error(err)
	}
}
