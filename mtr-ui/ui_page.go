package main

import (
	"github.com/GeoNet/mtr/mtrpb"
	"net/url"
)

type mtrUiPage struct {
	page
	Panels        []panel
	SparkGroups   []sparkGroup
	GroupRows     []idCount
	AppIDs        []app
	Path          string
	ModelID       string
	SiteID        string
	DeviceID      string
	TypeID        string
	ApplicationID string
	Status        string
	MtrApiUrl     string
	Resolution    string
	fieldResult   []*mtrpb.FieldMetricSummary
	dataResult    []*mtrpb.DataLatencySummary
	param         string
}

type panel struct {
	ID         string
	Title      string
	StatusLink string
	Values     map[string]idCount
	devices    map[string]bool
	sites      map[string]bool
}

type sparkGroup struct {
	ID    string
	Title string
	Rows  []sparkRow
}

type sparkRow struct {
	ID       string
	Title    string
	SparkUrl string
	Link     string
	Status   string
}

type idCount struct {
	ID          string
	Description string
	Link        string
	Count       int
}

type app struct {
	ID string
}

type panels []panel
type sparkRows []sparkRow
type sparkGroups []sparkGroup
type idCounts []idCount

func incCount(m map[string]idCount, key string) {
	t := m[key]
	t.Count = t.Count + 1
	m[key] = t
}

func (p *mtrUiPage) pageParam(q url.Values) int {
	n := 0
	p.param = ""
	p.ModelID = q.Get("modelID")
	if p.ModelID != "" {
		n++
		p.param = "modelID=" + p.ModelID
	}
	p.TypeID = q.Get("typeID")
	if p.TypeID != "" {
		n++
		if n > 1 {
			p.param = p.param + "&"
		}
		p.param = p.param + "typeID=" + p.TypeID
	}
	p.DeviceID = q.Get("deviceID")
	if p.DeviceID != "" {
		n++
		if n > 1 {
			p.param = p.param + "&"
		}
		p.param = p.param + "deviceID=" + p.DeviceID
	}
	p.SiteID = q.Get("siteID")
	if p.SiteID != "" {
		n++
		if n > 1 {
			p.param = p.param + "&"
		}
		p.param = p.param + "siteID=" + p.SiteID
	}
	p.ApplicationID = q.Get("applicationID")
	if p.ApplicationID != "" {
		n++
		if n > 1 {
			p.param = p.param + "&"
		}
		p.param = p.param + "applicationID=" + p.ApplicationID
	}
	p.Status = q.Get("status")
	if p.Status != "" {
		n++
		if n > 1 {
			p.param = p.param + "&"
		}
		p.param = p.param + "status=" + p.Status
	}

	p.Resolution = q.Get("resolution")
	return n
}

func (p mtrUiPage) appendPageParam(s string) string {
	if p.param != "" {
		s = s + "&" + p.param
	}
	return s
}

func (m panels) Len() int {
	return len(m)
}

func (m panels) Less(i, j int) bool {
	return m[i].Title < m[j].Title
}

func (m panels) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m sparkRows) Len() int {
	return len(m)
}

func (m sparkRows) Less(i, j int) bool {
	return m[i].Title < m[j].Title
}

func (m sparkRows) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m sparkGroups) Len() int {
	return len(m)
}

func (m sparkGroups) Less(i, j int) bool {
	return m[i].Title < m[j].Title
}

func (m sparkGroups) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m idCounts) Len() int {
	return len(m)
}

func (m idCounts) Less(i, j int) bool {
	return m[i].ID < m[j].ID
}

func (m idCounts) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
