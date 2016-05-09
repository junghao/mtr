package main

import (
	"bytes"
	"net/http"
)

type homepage struct {
	page
	FieldSummary map[string]int
	DataSummary  map[string]int
	AppSummary   map[string]int
}

func homepageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {

	var err error

	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	p := homepage{}
	p.Border.Title = "GeoNet MTR"

	if err = p.populateTags(); err != nil {
		return internalServerError(err)
	}

	if p.FieldSummary, err = getFieldSummary(); err != nil {
		return internalServerError(err)
	}

	if p.DataSummary, err = getDataSummary(); err != nil {
		return internalServerError(err)
	}

	if err = homepageTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}
