package main

import (
	"bytes"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"net/http"
	"sort"
)

type fieldPage struct {
	page
	Path         string
	Summary      map[string]int
	Metrics      []idCount
	DeviceModels []deviceModel
	Devices      []device
	ModelID      string
	DeviceID     string
	TypeID       string
	Status       string
	Resolution   string
	MtrApiUrl    string
}

type deviceModels []deviceModel

func (m deviceModels) Len() int {
	return len(m)
}

func (m deviceModels) Less(i, j int) bool {
	return m[i].ModelID < m[j].ModelID
}

func (m deviceModels) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type devices []device

func (m devices) Len() int {
	return len(m)
}

func (m devices) Less(i, j int) bool {
	return m[i].DeviceID < m[j].DeviceID
}

func (m devices) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type idCounts []idCount

func (m idCounts) Len() int {
	return len(m)
}

func (m idCounts) Less(i, j int) bool {
	return m[i].ID < m[j].ID
}

func (m idCounts) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type deviceModel struct {
	ModelID     string
	TypeCount   int
	DeviceCount int
	Count       map[string]int
}

type device struct {
	DeviceID string
	ModelID  string
	typeStatus
}

type typeStatus struct {
	TypeID string
	Status string
}

type idCount struct {
	ID          string
	Description string
	Count       map[string]int
}

func fieldPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {

	var err error

	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	p := fieldPage{}
	p.Path = r.URL.Path
	p.Border.Title = "GeoNet MTR"

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	if p.Summary, err = getFieldSummary(); err != nil {
		return weft.InternalServerError(err)
	}

	if err = fieldTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func fieldMetricsPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {

	var err error

	if res := weft.CheckQuery(r, []string{}, []string{"status"}); !res.Ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	p := fieldPage{}
	p.Path = r.URL.Path
	p.Border.Title = "GeoNet MTR"
	p.MtrApiUrl = mtrApiUrl.String()
	q := r.URL.Query()
	p.Status = q.Get("status")

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	if p.Status != "" {
		if err = p.getDevicesByStatus(); err != nil {
			return weft.InternalServerError(err)
		}
	} else {
		if err = p.getMetricsSummary(); err != nil {
			return weft.InternalServerError(err)
		}
	}

	if err = fieldTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func fieldDevicesPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {

	var err error

	if res := weft.CheckQuery(r, []string{}, []string{"modelID", "typeID", "status"}); !res.Ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	p := fieldPage{}
	p.Path = r.URL.Path
	p.MtrApiUrl = mtrApiUrl.String()
	p.Border.Title = "GeoNet MTR"

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	q := r.URL.Query()
	p.ModelID = q.Get("modelID")
	p.TypeID = q.Get("typeID")
	p.Status = q.Get("status")
	if p.ModelID != "" && p.TypeID != "" {
		if err = p.getDevicesByModelType(); err != nil {
			return weft.InternalServerError(err)
		}
	} else if p.ModelID != "" && p.Status != "" {
		if err = p.getDevicesByModelStatus(); err != nil {
			return weft.InternalServerError(err)
		}
	} else {
		if err = p.getDevicesSummary(); err != nil {
			return weft.InternalServerError(err)
		}
	}
	if err = fieldTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func fieldPlotPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "typeID"}, []string{"resolution"}); !res.Ok {
		return res
	}
	p := fieldPage{}
	p.Path = r.URL.Path
	p.MtrApiUrl = mtrApiUrl.String()
	p.Border.Title = "GeoNet MTR"
	q := r.URL.Query()
	p.DeviceID = q.Get("deviceID")
	p.TypeID = q.Get("typeID")
	p.Resolution = q.Get("resolution")
	if p.Resolution == "" {
		p.Resolution = "minute"
	}

	if err := fieldTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func getFieldSummary() (m map[string]int, err error) {
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

	m = make(map[string]int)
	m["metrics"] = len(f.Result)
	devices := make(map[string]bool)
	for _, r := range f.Result {
		devices[r.DeviceID] = true
		incFieldCount(m, r)
	}
	m["devices"] = len(devices)
	return
}

func (p *fieldPage) getMetricsSummary() (err error) {
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

	p.Metrics = make([]idCount, 0)
	for _, r := range f.Result {
		p.Metrics = updateFieldMetric(p.Metrics, r)
	}

	sort.Sort(idCounts(p.Metrics))
	return
}

func (p *fieldPage) getDevicesSummary() (err error) {
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

	p.DeviceModels = make([]deviceModel, 0)
	for _, r := range f.Result {
		p.DeviceModels = updateFieldDevice(p.DeviceModels, r)
	}

	sort.Sort(deviceModels(p.DeviceModels))
	return
}

func (p *fieldPage) getDevicesByModelStatus() (err error) {
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

	p.Devices = make([]device, 0)
	for _, r := range f.Result {
		if r.ModelID == p.ModelID && fieldStatusString(r) == p.Status {
			t := device{ModelID: p.ModelID, DeviceID: r.DeviceID}
			t.TypeID = r.TypeID
			t.Status = fieldStatusString(r)
			p.Devices = append(p.Devices, t)
		}
	}

	sort.Sort(devices(p.Devices))
	return
}

func (p *fieldPage) getDevicesByModelType() (err error) {
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

	p.Devices = make([]device, 0)
	for _, r := range f.Result {
		if r.ModelID == p.ModelID && r.TypeID == p.TypeID {
			t := device{ModelID: p.ModelID, DeviceID: r.DeviceID}
			t.TypeID = r.TypeID
			t.Status = fieldStatusString(r)
			p.Devices = append(p.Devices, t)
		}
	}

	sort.Sort(devices(p.Devices))
	return
}

func (p *fieldPage) getDevicesByStatus() (err error) {
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

	p.Devices = make([]device, 0)
	for _, r := range f.Result {
		if fieldStatusString(r) == p.Status {
			t := device{ModelID: r.ModelID, DeviceID: r.DeviceID}
			t.TypeID = r.TypeID
			t.Status = fieldStatusString(r)
			p.Devices = append(p.Devices, t)
		}
	}

	sort.Sort(devices(p.Devices))
	return
}

// Increase count if ID exists in slice, append to slice if it's a new ID
func updateFieldMetric(m []idCount, result *mtrpb.FieldMetricSummary) []idCount {
	for _, r := range m {
		if r.ID == result.TypeID {
			incFieldCount(r.Count, result)
			return m
		}
	}

	c := make(map[string]int)
	incFieldCount(c, result)
	return append(m, idCount{ID: result.TypeID, Count: c})
}

// Increase count if ID exists in slice, append to slice if it's a new ID
func updateFieldDevice(m []deviceModel, result *mtrpb.FieldMetricSummary) []deviceModel {
	for i, r := range m {
		if r.ModelID == result.ModelID {
			r.TypeCount++
			incFieldCount(r.Count, result)
			m[i] = r
			return m
		}
	}

	c := make(map[string]int)
	incFieldCount(c, result)

	return append(m, deviceModel{ModelID: result.ModelID, Count: c, TypeCount: 1})
}

func incFieldCount(m map[string]int, r *mtrpb.FieldMetricSummary) {
	s := fieldStatusString(r)
	m[s] = m[s] + 1
	m["total"] = m["total"] + 1
}

func fieldStatusString(r *mtrpb.FieldMetricSummary) string {
	switch {
	case r.Upper == 0 && r.Lower == 0:
		return "unknown"
	case r.Value >= r.Lower && r.Value <= r.Upper:
		return "good"
		// TBD: late
	}
	return "bad"
}
