package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

type metricDetailPage struct {
	page
	MtrApiUrl    *url.URL
	MetricDetail metricDetail
}

type metricDetail struct {
	DeviceID string
	TypeID   string
}

// handler that serves an html page for detailed metric information
func metricDetailHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {

	var (
		err error
	)

	if res := checkQuery(r, []string{"deviceID", "typeID"}, []string{}); !res.ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	q := r.URL.Query()
	deviceID := q.Get("deviceID")
	typeID := q.Get("typeID")

	p := metricDetailPage{MtrApiUrl: mtrApiUrl}
	p.Border.Title = fmt.Sprintf("Detailed Metric Info for deviceID:%s TypeID:%s", deviceID, typeID)
	p.MetricDetail.DeviceID = deviceID
	p.MetricDetail.TypeID = typeID

	if err = p.populateTags(); err != nil {
		return internalServerError(err)
	}

	if err = metricDetailTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}
