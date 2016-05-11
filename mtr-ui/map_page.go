package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

type mapPage struct {
	page
	MtrApiUrl string
	TypeID    string
}

func mapPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {

	var err error

	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	p := mapPage{}
	p.MtrApiUrl = mtrApiUrl.String()
	p.Border.Title = "GeoNet MTR"

	if err = p.populateTags(); err != nil {
		return internalServerError(err)
	}

	s := strings.TrimPrefix(r.URL.Path, "/map/")
	switch s {
	case "":
		p.TypeID = "voltage"
	case "voltage", "conn", "ping":
		p.TypeID = s
	default:
		return internalServerError(fmt.Errorf("Unknown map type"))
	}

	if err = mapTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}
