package main

import (
	"bytes"
	"fmt"
	"github.com/GeoNet/weft"
	"net/http"
	"strings"
)

type mapPage struct {
	page
	MtrApiUrl string
	TypeID    string
}

func mapPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {

	var err error

	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	p := mapPage{}
	p.MtrApiUrl = mtrApiUrl.String()
	p.Border.Title = "GeoNet MTR"

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	s := strings.TrimPrefix(r.URL.Path, "/map")
	switch s {
	case "", "/":
		p.TypeID = "voltage"
	case "/voltage", "/conn", "/ping":
		p.TypeID = s
	default:
		return weft.InternalServerError(fmt.Errorf("Unknown map type"))
	}

	if err = mapTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
