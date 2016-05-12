package main

import (
	"bytes"
	"github.com/GeoNet/weft"
	"net/http"
)

type homepage struct {
	page
	FieldSummary map[string]int
	DataSummary  map[string]int
	AppSummary   map[string]int
}

func homepageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {

	var err error

	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	p := homepage{}
	p.Border.Title = "GeoNet MTR"

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	if p.FieldSummary, err = getFieldSummary(); err != nil {
		return weft.InternalServerError(err)
	}

	if p.DataSummary, err = getDataSummary(); err != nil {
		return weft.InternalServerError(err)
	}

	if err = homepageTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
