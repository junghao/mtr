package main

import (
	"bytes"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"net/http"
	"sort"
)

func fieldPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error

	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	p := mtrUiPage{}
	p.Path = r.URL.Path
	p.Border.Title = "GeoNet MTR"
	p.ActiveTab = "Field"

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	var pa panel
	if pa, err = getFieldSummary(); err != nil {
		return weft.InternalServerError(err)
	}

	p.Panels = []panel{pa}

	if err = fieldTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func fieldMetricsPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error
	if res := weft.CheckQuery(r, []string{}, []string{"status", "modelID", "typeID", "deviceID"}); !res.Ok {
		return res
	}

	p := mtrUiPage{}
	p.Path = r.URL.Path
	p.MtrApiUrl = mtrApiUrl.String()
	p.Border.Title = "GeoNet MTR"
	p.ActiveTab = "Field"
	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	n := p.pageParam(r.URL.Query())

	// For /field/metrics, we :
	// 1. Show grouped list if only Status parameter is specified,
	// 2. Show list when Status or ModelID parameter is specified,
	// Else we show panel.
	if n == 1 && p.Status != "" {
		if err = p.getFieldCountList(); err != nil {
			return weft.InternalServerError(err)
		}
	} else if p.Status != "" || p.ModelID != "" {
		if err = p.getDevicesList(); err != nil {
			return weft.InternalServerError(err)
		}
	} else {
		if err = p.getFieldMetricsPanel(); err != nil {
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

	if res := weft.CheckQuery(r, []string{}, []string{"status", "modelID", "typeID", "deviceID"}); !res.Ok {
		return res
	}

	p := mtrUiPage{}
	p.Path = r.URL.Path
	p.MtrApiUrl = mtrApiUrl.String()
	p.Border.Title = "GeoNet MTR"
	p.ActiveTab = "Field"

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	p.pageParam(r.URL.Query())

	// For /field/devices, we show list when Status or ModelID parameter is specified.
	// Else we show panel.
	if p.Status != "" || p.ModelID != "" {
		if err = p.getDevicesList(); err != nil {
			return weft.InternalServerError(err)
		}
	} else {
		if err = p.getDevicesPanel(); err != nil {
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
	p := mtrUiPage{}
	p.Path = r.URL.Path
	p.MtrApiUrl = mtrApiUrl.String()
	p.Border.Title = "GeoNet MTR - Field"
	p.ActiveTab = "Field"
	p.pageParam(r.URL.Query())

	if err := p.getFieldMetricTags(); err != nil {
		return weft.InternalServerError(err)
	}

	if p.Resolution == "" {
		p.Resolution = "hour"
	}

	if err := p.getFieldHistoryLog(); err != nil {
		return weft.InternalServerError(err)
	}

	if err := p.getFieldYLabel(); err != nil {
		return weft.InternalServerError(err)
	}

	// Set thresholds on plot by drawing a box in dygraph.  Protobuf contains all thresholds, so select ours
	u := *mtrApiUrl
	u.Path = "/field/metric/threshold"

	var err error
	var protoData []byte
	if protoData, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return weft.InternalServerError(err)
	}

	var f mtrpb.FieldMetricThresholdResult
	if err = proto.Unmarshal(protoData, &f); err != nil {
		return weft.InternalServerError(err)
	}

	if f.Result != nil {
		for _, row := range f.Result {
			if row.DeviceID == p.DeviceID && row.TypeID == p.TypeID {
				p.Plt.Thresholds = []float64{float64(row.Lower) * row.Scale, float64(row.Upper) * row.Scale}
			}
		}
	}

	if err := fieldTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

// For home screen panel only
func getFieldSummary() (p panel, err error) {
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

	p.Title = "Fields"
	p.StatusLink = "/field/metrics?"
	m := make(map[string]idCount, 0)

	devices := make(map[string]bool)
	for _, r := range f.Result {
		devices[r.DeviceID] = true
		incFieldCount(m, r)
	}
	// Update header part of panel
	m["devices"] = idCount{Count: len(devices), ID: "Devices", Link: "/field/devices"}
	m["metrics"] = idCount{Count: len(f.Result), ID: "Metrics", Link: "/field/metrics"}
	p.Values = m
	return
}

// Path: /field/metrics
func (p *mtrUiPage) getFieldMetricsPanel() (err error) {
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

	p.Panels = make([]panel, 0)
	p.fieldResult = p.filterFieldResults(f.Result)

	for _, r := range p.fieldResult {
		if p.TypeID != "" {
			p.updateFieldMetricGroupByModel(r)
		} else {
			p.updateFieldMetricGroupByType(r)
		}
	}

	// Update header part of panel
	for _, r := range p.Panels {
		var dl, ml string
		if p.TypeID != "" {
			dl = p.appendPageParam("/field/devices?modelID=" + r.ID)
			ml = p.appendPageParam("/field/metrics?modelID=" + r.ID)
		} else {
			dl = p.appendPageParam("/field/devices?typeID=" + r.ID)
			ml = p.appendPageParam("/field/metrics?typeID=" + r.ID)
		}
		m := idCount{Count: len(r.devices), ID: "Devices", Link: dl}
		r.Values["devices"] = m
		m = idCount{Count: r.Values["total"].Count, ID: "Metrics", Link: ml}
		r.Values["metrics"] = m
	}
	sort.Sort(panels(p.Panels))
	return
}

// Path: /field/devices
func (p *mtrUiPage) getDevicesPanel() (err error) {
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

	p.Panels = make([]panel, 0)
	p.fieldResult = p.filterFieldResults(f.Result)

	for _, r := range p.fieldResult {
		p.updateFieldDevice(r)
	}

	// Update header part of panel
	for _, r := range p.Panels {
		if p.ModelID == "" {
			l := p.appendPageParam("/field/devices?modelID=" + r.ID)
			m := idCount{Count: len(r.devices), ID: "Devices", Link: l}
			r.Values["devices"] = m
		}
		if p.TypeID == "" {
			l := p.appendPageParam("/field/metrics?modelID=" + r.ID)
			m := idCount{Count: r.Values["total"].Count, ID: "Metrics", Link: l}
			r.Values["metrics"] = m
		}
	}
	sort.Sort(panels(p.Panels))
	return
}

func (p *mtrUiPage) getDevicesList() (err error) {
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

	p.SparkGroups = make([]sparkGroup, 0)
	p.fieldResult = p.filterFieldResults(f.Result)

	// We don't aggregate if both modelID and typeID are specified.
	// Default to group by ModelID. (If both TypeID and ModelID are missing.)
	groupByModel := true
	aggr := (p.ModelID == "" || p.TypeID == "")
	if !aggr {
		if len(p.fieldResult) > 0 {
			p.SparkGroups = append(p.SparkGroups, sparkGroup{Rows: make([]sparkRow, 0)})
		}
	} else {
		if p.ModelID != "" {
			groupByModel = false
		}
	}

	for _, r := range p.fieldResult {
		s := fieldStatusString(r)
		row := sparkRow{
			ID:       r.DeviceID + " " + r.TypeID,
			Title:    r.DeviceID + " " + r.TypeID,
			Link:     "/field/plot?deviceID=" + r.DeviceID + "&typeID=" + r.TypeID,
			SparkUrl: "/field/metric?deviceID=" + r.DeviceID + "&typeID=" + r.TypeID,
			Status:   s,
		}

		stored := false
		for i, g := range p.SparkGroups {
			// If we're not doing aggregation then we always add new row into first group
			if !aggr || (!groupByModel && g.ID == r.TypeID) || (groupByModel && g.ID == r.ModelID) {
				g.Rows = append(g.Rows, row)
				p.SparkGroups[i] = g
				stored = true
				break
			}
		}
		if stored {
			continue
		}
		// Cannot find a matching group, create a new group
		var sg sparkGroup
		if groupByModel {
			sg = sparkGroup{ID: r.ModelID, Title: r.ModelID, Rows: []sparkRow{row}}
		} else {
			sg = sparkGroup{ID: r.TypeID, Title: r.TypeID, Rows: []sparkRow{row}}
		}
		p.SparkGroups = append(p.SparkGroups, sg)

	}

	for i, g := range p.SparkGroups {
		sort.Sort(sparkRows(g.Rows))
		p.SparkGroups[i] = g
	}
	sort.Sort(sparkGroups(p.SparkGroups))
	return
}

// getFieldCountList returns []idCount for each typeID
func (p *mtrUiPage) getFieldCountList() (err error) {
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

	// The trick here is to create panels first -
	//   then use aggregate functions for panels,
	//   then transfer from panels to idCounts
	p.Panels = make([]panel, 0)
	p.fieldResult = p.filterFieldResults(f.Result)

	for _, r := range p.fieldResult {
		if p.TypeID != "" {
			p.updateFieldMetricGroupByModel(r)
		} else {
			p.updateFieldMetricGroupByType(r)
		}
	}

	// Now copy counts from panels to GroupRows
	p.GroupRows = make([]idCount, 0)
	for _, r := range p.Panels {
		// Note: getCountList only count for same Status
		c := idCount{ID: r.ID, Description: r.ID, Link: r.StatusLink, Count: r.Values[p.Status].Count}
		p.GroupRows = append(p.GroupRows, c)
	}
	sort.Sort(idCounts(p.GroupRows))
	p.Panels = nil
	return
}

func (p *mtrUiPage) getFieldMetricTags() (err error) {
	u := *mtrApiUrl
	u.Path = "/field/metric/tag"
	u.RawQuery = "deviceID=" + p.DeviceID + "&typeID=" + p.TypeID

	var b []byte
	if b, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return
	}

	var f mtrpb.FieldMetricTagResult

	if err = proto.Unmarshal(b, &f); err != nil {
		return
	}

	for _, r := range f.Result {
		p.Tags = append(p.Tags, r.Tag)
	}

	sort.Strings(p.Tags)
	return
}

func (p *mtrUiPage) getFieldHistoryLog() (err error) {
	u := *mtrApiUrl
	u.Path = "/field/metric"
	u.RawQuery = "deviceID=" + p.DeviceID + "&typeID=" + p.TypeID + "&resolution=" + p.Resolution
	var b []byte
	if b, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return
	}

	var f mtrpb.FieldMetricResult

	if err = proto.Unmarshal(b, &f); err != nil {
		return
	}

	p.FieldLog = &f
	return
}

func (p *mtrUiPage) getFieldYLabel() (err error) {
	u := *mtrApiUrl
	u.Path = "/field/type"
	var b []byte
	if b, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return
	}

	var f mtrpb.FieldTypeResult

	if err = proto.Unmarshal(b, &f); err != nil {
		return
	}

	for _, val := range f.Result {
		if val.TypeID == p.TypeID {
			p.Plt.Ylabel = val.Display
			return
		}
	}

	return
}

func (p mtrUiPage) filterFieldResults(f []*mtrpb.FieldMetricSummary) []*mtrpb.FieldMetricSummary {
	result := make([]*mtrpb.FieldMetricSummary, 0)

	for _, r := range f {
		if p.ModelID != "" && p.ModelID != r.ModelID {
			continue
		}
		if p.Status != "" && p.Status != fieldStatusString(r) {
			continue
		}
		if p.TypeID != "" && p.TypeID != r.TypeID {
			continue
		}
		if p.DeviceID != "" && p.DeviceID != r.DeviceID {
			continue
		}
		result = append(result, r)
	}

	return result
}

// Increase count if ID exists in slice, append to slice if it's a new ID
func (p *mtrUiPage) updateFieldMetricGroupByType(result *mtrpb.FieldMetricSummary) {
	for i, r := range p.Panels {
		if r.ID == result.TypeID {
			r.devices[result.DeviceID] = true
			incFieldCount(r.Values, result)
			p.Panels[i] = r
			return
		}
	}

	c := make(map[string]idCount)
	incFieldCount(c, result)

	d := make(map[string]bool)
	d[result.DeviceID] = true

	l := p.appendPageParam(p.Path + "?typeID=" + result.TypeID)
	p.Panels = append(p.Panels, panel{ID: result.TypeID, Title: result.TypeID, StatusLink: l, Values: c, devices: d})
}

func (p *mtrUiPage) updateFieldMetricGroupByModel(result *mtrpb.FieldMetricSummary) {
	for i, r := range p.Panels {
		if r.ID == result.ModelID {
			r.devices[result.DeviceID] = true
			incFieldCount(r.Values, result)
			p.Panels[i] = r
			return
		}
	}

	c := make(map[string]idCount)
	incFieldCount(c, result)

	d := make(map[string]bool)
	d[result.DeviceID] = true

	l := p.appendPageParam(p.Path + "?modelID=" + result.ModelID)
	p.Panels = append(p.Panels, panel{ID: result.ModelID, Title: result.ModelID, StatusLink: l, Values: c, devices: d})

}

func (p *mtrUiPage) updateFieldDevice(result *mtrpb.FieldMetricSummary) {
	// updateFieldDevice does the same thing as updateFieldMetricGroupByModel
	p.updateFieldMetricGroupByModel(result)
}

func incFieldCount(m map[string]idCount, r *mtrpb.FieldMetricSummary) {
	s := fieldStatusString(r)
	incCount(m, s)
	incCount(m, "total")
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
