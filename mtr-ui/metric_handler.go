package main

import (
	"bytes"
	"fmt"
	"github.com/GeoNet/weft"
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
func metricDetailHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {

	var (
		err error
	)

	if res := weft.CheckQuery(r, []string{"deviceID", "typeID"}, []string{}); !res.Ok {
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
		return weft.InternalServerError(err)
	}

	if err = metricDetailTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
