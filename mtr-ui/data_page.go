package main

import (
	"bytes"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/golang/protobuf/proto"
	"net/http"
)

type dataPage struct {
	page
	Path    string
	Summary map[string]int
}

func dataPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {

	var err error

	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	p := dataPage{}
	p.Border.Title = "GeoNet MTR"

	if err = p.populateTags(); err != nil {
		return internalServerError(err)
	}

	if p.Summary, err = getDataSummary(); err != nil {
		return internalServerError(err)
	}

	p.Path = r.URL.Path
	if err = fieldTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func getDataSummary() (m map[string]int, err error) {
	u := *mtrApiUrl
	u.Path = "/field/metric/summary"

	var b []byte
	if b, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return
	}

	var f mtrpb.FieldMetricSummaryResult

	if err = proto.Unmarshal(b, &f); err != nil {
		return
	}

	m = make(map[string]int, 6)
	m["metrics"] = len(f.Result)
	devices := make(map[string]bool)
	for _, r := range f.Result {
		devices[r.DeviceID] = true
		incCount(m, r)
	}
	m["devices"] = len(devices)
	return
}
