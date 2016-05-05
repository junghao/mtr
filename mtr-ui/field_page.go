package main

import (
	"bytes"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/golang/protobuf/proto"
	"net/http"
	"sort"
)

type fieldPage struct {
	page
	Path    string
	Summary map[string]int
	Metrics []idCount
	Devices []device
}

type devices []device

func (m devices) Len() int {
	return len(m)
}

func (m devices) Less(i, j int) bool {
	return m[i].ModelId < m[j].ModelId
}

func (m devices) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type idCounts []idCount

func (m idCounts) Len() int {
	return len(m)
}

func (m idCounts) Less(i, j int) bool {
	return m[i].Id < m[j].Id
}

func (m idCounts) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type device struct {
	ModelId   string
	TypeCount int
	Types     []idCount
}

type idCount struct {
	Id    string
	Count map[string]int
}

func fieldPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {

	var err error

	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	p := fieldPage{}
	p.Path = r.URL.Path
	p.Border.Title = "GeoNet MTR"

	if err = p.populateTags(); err != nil {
		return internalServerError(err)
	}

	if p.Summary, err = getFieldLatest(); err != nil {
		return internalServerError(err)
	}

	if err = fieldTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func fieldMetricsPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {

	var err error

	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	p := fieldPage{}
	p.Path = r.URL.Path
	p.Border.Title = "GeoNet MTR"

	if err = p.populateTags(); err != nil {
		return internalServerError(err)
	}

	if err = p.getMetricsLatest(); err != nil {
		return internalServerError(err)
	}

	if err = fieldTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func fieldDevicesPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {

	var err error

	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	p := fieldPage{}
	p.Path = r.URL.Path
	p.Border.Title = "GeoNet MTR"

	if err = p.populateTags(); err != nil {
		return internalServerError(err)
	}

	if err = p.getDevicesLatest(); err != nil {
		return internalServerError(err)
	}

	if err = fieldTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func getFieldLatest() (m map[string]int, err error) {
	u := *mtrApiUrl
	u.Path = "/field/metric/latest"

	var b []byte
	if b, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return
	}

	var f mtrpb.FieldMetricLatestResult

	if err = proto.Unmarshal(b, &f); err != nil {
		return
	}

	m = make(map[string]int)
	m["metrics"] = len(f.Result)
	devices := make(map[string]bool)
	for _, r := range f.Result {
		devices[r.DeviceID] = true
		incCount(m, r)
	}
	m["devices"] = len(devices)
	return
}

func (p *fieldPage) getMetricsLatest() (err error) {
	u := *mtrApiUrl
	u.Path = "/field/metric/latest"

	var b []byte
	if b, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return
	}

	var f mtrpb.FieldMetricLatestResult

	if err = proto.Unmarshal(b, &f); err != nil {
		return
	}

	p.Metrics = make([]idCount, 0)
	for _, r := range f.Result {
		p.Metrics = updateMetric(p.Metrics, r)
	}

	sort.Sort(idCounts(p.Metrics))
	return
}

func (p *fieldPage) getDevicesLatest() (err error) {
	u := *mtrApiUrl
	u.Path = "/field/metric/latest"

	var b []byte
	if b, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return
	}

	var f mtrpb.FieldMetricLatestResult

	if err = proto.Unmarshal(b, &f); err != nil {
		return
	}

	p.Devices = make([]device, 0)
	for _, r := range f.Result {
		p.Devices = updateDevice(p.Devices, r)
	}

	sort.Sort(devices(p.Devices))
	return
}

// Increase count if Id exists in slice, append to slice if it's a new Id
func updateMetric(m []idCount, result *mtrpb.FieldMetricLatest) []idCount {
	for _, r := range m {
		if r.Id == result.TypeID {
			incCount(r.Count, result)
			return m
		}
	}

	c := make(map[string]int)
	incCount(c, result)
	return append(m, idCount{Id: result.TypeID, Count: c})
}

// Increase count if Id exists in slice, append to slice if it's a new Id
func updateDevice(m []device, result *mtrpb.FieldMetricLatest) []device {
	for i, r := range m {
		if r.ModelId == result.ModelID {
			r.TypeCount++
			for j, rt := range r.Types {
				if rt.Id == result.TypeID {
					incCount(rt.Count, result)
					r.Types[j] = rt
					m[i] = r
					return m
				}
			}
			// create a new typeId in this modelId
			c := make([]idCount, 0)
			r.Types = updateMetric(c, result)
			m[i] = r
			return m
		}
	}

	c := make(map[string]int)
	incCount(c, result)

	t := []idCount{idCount{Id: result.TypeID, Count: c}}
	return append(m, device{ModelId: result.ModelID, Types: t, TypeCount: 1})
}

func incCount(m map[string]int, r *mtrpb.FieldMetricLatest) {
	switch {
	case r.Upper == 0 && r.Lower == 0:
		m["unknown"] = m["unknown"] + 1
	case r.Value >= r.Lower && r.Value <= r.Upper:
		m["good"] = m["good"] + 1
	default:
		m["bad"] = m["bad"] + 1
		// TBD: late
	}
	m["total"] = m["total"] + 1
}
