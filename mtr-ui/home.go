package main

import (
	"bytes"
	"github.com/GeoNet/weft"
	"net/http"
)

func homePageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {

	var err error

	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	p := mtrUiPage{}
	p.Border.Title = "GeoNet MTR"

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	var fieldPanel, dataPanel panel

	if fieldPanel, err = getFieldSummary(); err != nil {
		return weft.InternalServerError(err)
	}

	if dataPanel, err = getDataSummary(); err != nil {
		return weft.InternalServerError(err)
	}

	p.Panels = []panel{dataPanel, fieldPanel}
	if err = homepageTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
