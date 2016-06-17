package main

import (
	"bytes"
	"fmt"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"net/http"
	"net/url"
)

type searchPage struct {
	page
	ActiveTab       string // used to satisfy the templates, but not used for search page
	MtrApiUrl       *url.URL
	TagName         string
	MatchingMetrics matchingMetrics
}

type matchingMetrics []metricInfo

type metricInfo struct {
	TypeID           string
	DeviceID         string
	SiteID           string
	Tag              string
	Status           string
	CompletenessInfo string
}

func newSearchPage(apiUrl *url.URL) (s *searchPage, err error) {
	s = &searchPage{MtrApiUrl: apiUrl}
	s.Border.Title = "GeoNet MTR - Search Results"
	return s, nil
}

func (s *searchPage) matchingMetrics(tagQuery string) (err error) {
	u := *s.MtrApiUrl
	u.Path = "/tag/" + tagQuery
	if s.MatchingMetrics, err = getMatchingMetrics(u.String()); err != nil {
		return err
	}
	s.TagName = tagQuery

	return nil
}

func getMatchingMetrics(urlString string) (parsedTags matchingMetrics, err error) {

	b, err := getBytes(urlString, "application/x-protobuf")
	if err != nil {
		return nil, err
	}

	var tr mtrpb.TagSearchResult

	if err = proto.Unmarshal(b, &tr); err != nil {
		return nil, err
	}

	if tr.FieldMetric != nil {
		for _, v := range tr.FieldMetric {
			m := metricInfo{
				TypeID:   v.TypeID,
				DeviceID: v.DeviceID,
				Status:   fieldStatusString(v),
			}
			parsedTags = append(parsedTags, m)
		}
	}

	if tr.DataLatency != nil {
		for _, v := range tr.DataLatency {
			m := metricInfo{
				TypeID: v.TypeID,
				SiteID: v.SiteID,
				Status: dataStatusString(v),
			}
			parsedTags = append(parsedTags, m)
		}
	}

	if tr.DataCompleteness != nil {
		for _, v := range tr.DataCompleteness {
			// Data completeness default returns
			m := metricInfo{
				TypeID:           v.TypeID,
				SiteID:           v.SiteID,
				Status:           completenessStatusString(v),
				CompletenessInfo: fmt.Sprintf("%4.2f", v.Completeness),
			}
			parsedTags = append(parsedTags, m)
		}
	}

	return parsedTags, nil
}

func searchPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {

	var err error
	var p *searchPage

	if res := weft.CheckQuery(r, []string{"tagQuery"}, []string{"page"}); !res.Ok {
		return res
	}
	r.ParseForm()
	tagQuery := r.FormValue("tagQuery")

	// Javascript should handle empty query value
	// Non existent value comes from unauthorized submit
	if tagQuery == "" {
		return weft.BadRequest("missing required query parameter: tagQuery")
	}

	if p, err = newSearchPage(mtrApiUrl); err != nil {
		return weft.BadRequest("error creating searchPage object")
	}

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	if err = p.matchingMetrics(tagQuery); err != nil {
		return weft.InternalServerError(err)
	}

	if err := tagSearchTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
