package main

import (
	"bytes"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"net/http"
	"sort"
	"strings"
)

func dataPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error

	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	p := mtrUiPage{}
	p.Path = r.URL.Path
	p.Border.Title = "GeoNet MTR - Data"
	p.ActiveTab = "Data"

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	var pa panel
	if pa, err = getDataSummary(); err != nil {
		return weft.InternalServerError(err)
	}

	p.Panels = []panel{pa}

	if err = dataTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func dataMetricsPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error

	if res := weft.CheckQuery(r, []string{}, []string{"status", "typeID"}); !res.Ok {
		return res
	}

	p := mtrUiPage{}
	p.Path = r.URL.Path
	p.Border.Title = "GeoNet MTR - Data Metrics"
	p.ActiveTab = "Data"
	p.MtrApiUrl = mtrApiUrl.String()

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	n := p.pageParam(r.URL.Query())

	// For /data/metrics, we :
	// 1. Show grouped list if only Status parameter is specified,
	// 2. Show list when Status or TypeID parameter is specified,
	// Else we show panel.
	if n == 1 && p.Status != "" {
		if err = p.getDataCountList(); err != nil {
			return weft.InternalServerError(err)
		}
	} else if p.Status != "" || p.TypeID != "" {
		if err = p.getSitesList(); err != nil {
			return weft.InternalServerError(err)
		}
	} else {
		if err = p.getDataMetricsPanel(); err != nil {
			return weft.InternalServerError(err)
		}
	}

	if err = dataTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func dataSitesPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {

	var err error

	if res := weft.CheckQuery(r, []string{}, []string{"status", "typeID"}); !res.Ok {
		return res
	}

	p := mtrUiPage{}
	p.Path = r.URL.Path
	p.Border.Title = "GeoNet MTR - Data Sites"
	p.ActiveTab = "Data"
	p.MtrApiUrl = mtrApiUrl.String()

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	p.pageParam(r.URL.Query())

	if err = p.getSitesList(); err != nil {
		return weft.InternalServerError(err)
	}

	if err = dataTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func dataPlotPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{"resolution"}); !res.Ok {
		return res
	}
	p := mtrUiPage{}
	p.Path = r.URL.Path
	p.MtrApiUrl = mtrApiUrl.String()
	p.Border.Title = "GeoNet MTR - Data"
	p.ActiveTab = "Data"
	p.pageParam(r.URL.Query())

	if p.Resolution == "" {
		p.Resolution = "minute"
	}

	if err := dataTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func getDataSummary() (p panel, err error) {
	u := *mtrApiUrl
	u.Path = "/data/latency/summary"

	var b []byte
	if b, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return
	}

	var f mtrpb.DataLatencySummaryResult

	if err = proto.Unmarshal(b, &f); err != nil {
		return
	}

	p.Title = "Data"
	p.StatusLink = "/data/metrics?"
	m := make(map[string]idCount, 0)

	sites := make(map[string]bool)
	for _, r := range f.Result {
		sites[r.SiteID] = true
		incDataCount(m, r)
	}

	m["sites"] = idCount{Count: len(sites), ID: "Sites", Link: "/data/sites"}
	m["metrics"] = idCount{Count: len(f.Result), ID: "Metrics", Link: "/data/metrics"}
	p.Values = m

	return
}

func (p *mtrUiPage) getDataMetricsPanel() (err error) {
	u := *mtrApiUrl
	u.Path = "/data/latency/summary"

	var b []byte
	if b, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return
	}

	var f mtrpb.DataLatencySummaryResult

	if err = proto.Unmarshal(b, &f); err != nil {
		return
	}

	p.Panels = make([]panel, 0)
	p.dataResult = p.filterDataResults(f.Result)

	for _, r := range p.dataResult {
		p.updateDataMetric(r)
	}

	for _, r := range p.Panels {
		l := p.appendPageParam("/data/sites?typeID=" + r.ID)
		m := idCount{Count: len(r.devices), ID: "Sites", Link: l}
		r.Values["sites"] = m
		l = p.appendPageParam("/data/metrics?typeID=" + r.ID)
		m = idCount{Count: r.Values["total"].Count, ID: "Metrics", Link: l}
		r.Values["metrics"] = m
	}
	sort.Sort(panels(p.Panels))
	return
}

func (p *mtrUiPage) getSitesPanel() (err error) {
	u := *mtrApiUrl
	u.Path = "/data/latency/summary"

	var b []byte
	if b, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return
	}

	var f mtrpb.DataLatencySummaryResult

	if err = proto.Unmarshal(b, &f); err != nil {
		return
	}

	p.Panels = make([]panel, 0)
	p.dataResult = p.filterDataResults(f.Result)

	for _, r := range p.dataResult {
		p.updateDataSite(r)
	}

	for _, r := range p.Panels {
		if p.TypeID == "" {
			l := p.appendPageParam("/data/metrics?siteID=" + r.ID)
			m := idCount{Count: r.Values["total"].Count, ID: "Metrics", Link: l}
			r.Values["metrics"] = m
		}
	}
	sort.Sort(panels(p.Panels))
	return
}

func (p *mtrUiPage) getSitesList() (err error) {
	u := *mtrApiUrl
	u.Path = "/data/latency/summary"

	var b []byte
	if b, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return
	}

	var f mtrpb.DataLatencySummaryResult

	if err = proto.Unmarshal(b, &f); err != nil {
		return
	}

	p.SparkGroups = make([]sparkGroup, 0)
	p.dataResult = p.filterDataResults(f.Result)

	// We don't aggregate if typeID is specified
	if p.TypeID != "" && len(p.dataResult) > 0 {
		p.SparkGroups = append(p.SparkGroups, sparkGroup{Rows: make([]sparkRow, 0)})
	}

	for _, r := range p.dataResult {
		s := dataStatusString(r)
		row := sparkRow{
			ID:       r.SiteID + " " + r.TypeID,
			Title:    r.SiteID + " " + removeTypeIDPrefix(r.TypeID),
			Link:     "/data/plot?siteID=" + r.SiteID + "&typeID=" + r.TypeID,
			SparkUrl: "/data/latency?siteID=" + r.SiteID + "&typeID=" + r.TypeID,
			Status:   s,
		}

		stored := false
		for i, g := range p.SparkGroups {
			// If we're not doing aggregation(p.TypeID!="") then we always add new row into first group
			if p.TypeID != "" || g.ID == r.TypeID {
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
		sg = sparkGroup{ID: r.TypeID, Title: removeTypeIDPrefix(r.TypeID), Rows: []sparkRow{row}}
		p.SparkGroups = append(p.SparkGroups, sg)

	}

	for i, g := range p.SparkGroups {
		sort.Sort(sparkRows(g.Rows))
		p.SparkGroups[i] = g
	}
	sort.Sort(sparkGroups(p.SparkGroups))
	return
}

// getDataCountList returns []idCount for each typeID
func (p *mtrUiPage) getDataCountList() (err error) {
	u := *mtrApiUrl
	u.Path = "/data/latency/summary"

	var b []byte
	if b, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return
	}

	var f mtrpb.DataLatencySummaryResult

	if err = proto.Unmarshal(b, &f); err != nil {
		return
	}

	// The trick here is to create panels first -
	//   then use aggregate functions for panels,
	//   then transfer from panels to idCounts
	p.Panels = make([]panel, 0)
	p.dataResult = p.filterDataResults(f.Result)

	for _, r := range p.dataResult {
		p.updateDataMetric(r)
	}

	// Now copy counts from panels to GroupRows
	p.GroupRows = make([]idCount, 0)
	for _, r := range p.Panels {
		// Note: getCountList only count for same Status
		c := idCount{ID: r.ID, Description: removeTypeIDPrefix(r.ID), Link: r.StatusLink, Count: r.Values[p.Status].Count}
		p.GroupRows = append(p.GroupRows, c)
	}
	sort.Sort(idCounts(p.GroupRows))
	p.Panels = nil
	return
}

func (p mtrUiPage) filterDataResults(f []*mtrpb.DataLatencySummary) []*mtrpb.DataLatencySummary {
	result := make([]*mtrpb.DataLatencySummary, 0)

	for _, r := range f {
		if p.Status != "" && p.Status != dataStatusString(r) {
			continue
		}
		if p.TypeID != "" && p.TypeID != r.TypeID {
			continue
		}
		if p.SiteID != "" && p.SiteID != r.SiteID {
			continue
		}
		result = append(result, r)
	}

	return result
}

// Increase count if ID exists in slice, append to slice if it's a new ID
func (p *mtrUiPage) updateDataMetric(result *mtrpb.DataLatencySummary) {
	for i, r := range p.Panels {
		if r.ID == result.TypeID {
			r.devices[result.SiteID] = true
			incDataCount(r.Values, result)
			p.Panels[i] = r
			return
		}
	}

	c := make(map[string]idCount)
	incDataCount(c, result)

	d := make(map[string]bool)
	d[result.SiteID] = true

	l := p.appendPageParam(p.Path + "?typeID=" + result.TypeID)
	p.Panels = append(p.Panels, panel{ID: result.TypeID, Title: removeTypeIDPrefix(result.TypeID), StatusLink: l, Values: c, devices: d})
}

// Increase count if ID exists in slice, append to slice if it's a new ID
func (p *mtrUiPage) updateDataSite(result *mtrpb.DataLatencySummary) {
	for i, r := range p.Panels {
		if r.ID == result.SiteID {
			r.devices[result.SiteID] = true
			incDataCount(r.Values, result)
			p.Panels[i] = r
			return
		}
	}

	c := make(map[string]idCount)
	incDataCount(c, result)

	d := make(map[string]bool)
	d[result.SiteID] = true

	l := p.appendPageParam(p.Path + "?siteID=" + result.SiteID)
	p.Panels = append(p.Panels, panel{ID: result.TypeID, Title: result.SiteID, StatusLink: l, Values: c, devices: d})
}

func incDataCount(m map[string]idCount, r *mtrpb.DataLatencySummary) {
	s := dataStatusString(r)
	incCount(m, s)
	incCount(m, "total")
}

func dataStatusString(r *mtrpb.DataLatencySummary) string {
	switch {
	case r.Upper == 0 && r.Lower == 0:
		return "unknown"
	case allGood(r):
		return "good"
		// TBD: late
	}
	return "bad"
}

func allGood(r *mtrpb.DataLatencySummary) bool {
	if r.Upper == 0 && r.Lower == 0 {
		return false
	}
	if r.Mean < r.Lower || r.Mean > r.Upper {
		return false
	}
	if r.Fifty != 0 && (r.Fifty < r.Lower || r.Fifty > r.Upper) {
		return false
	}
	if r.Ninety != 0 && (r.Ninety < r.Lower || r.Ninety > r.Upper) {
		return false
	}
	return true
}

func removeTypeIDPrefix(typeID string) string {
	if strings.HasPrefix(typeID, "latency.") {
		return strings.TrimPrefix(typeID, "latency.")
	}

	return typeID
}
