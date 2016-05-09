package main

import (
	"bytes"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/golang/protobuf/proto"
	"net/http"
	"net/url"
)

type searchPage struct {
	page
	mtrApiUrl       *url.URL
	TagName         string
	MatchingMetrics matchingMetrics
}

type matchingMetrics []metricInfo

type metricInfo struct {
	TypeID   string
	DeviceID string
	Tag      string
	IconUrl  string
}

func newSearchPage(apiUrl *url.URL) (s *searchPage, err error) {
	s = &searchPage{mtrApiUrl: apiUrl}
	s.Border.Title = "GeoNet MTR - Search Results"
	return s, nil
}

func (s *searchPage) matchingMetrics(tagQuery string) (err error) {

	u := *s.mtrApiUrl
	u.Path = "/tag/" + tagQuery

	if s.MatchingMetrics, err = getMatchingMetrics(u.String()); err != nil {
		return err
	}

	// also keep track of the Tag ID we searched for
	s.TagName = tagQuery

	return nil
}

func (s *searchPage) fetchIcons() (err error) {

	u := *s.mtrApiUrl
	u.Path = "/field/metric"

	for idx, val := range s.MatchingMetrics {
		q := u.Query()
		q.Set("deviceID", val.DeviceID)
		q.Set("typeID", val.TypeID)
		q.Set("resolution", "hour")
		q.Set("plot", "spark")
		u.RawQuery = q.Encode()
		s.MatchingMetrics[idx].IconUrl = u.String()
	}

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
				// TODO - not sure what the Tag is
				//Tag      string
				// TODO - not sure what this is
				//IconUrl  string
			}
			parsedTags = append(parsedTags, m)
		}
	}

	// TODO add DataLatency.

	return parsedTags, nil
}

func searchHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {

	var err error
	var p *searchPage

	r.ParseForm()
	tagQuery := r.FormValue("tagQuery")

	// Javascript should handle empty query value
	// Non existent value comes from unauthorized submit
	if tagQuery == "" {
		return badRequest("missing required query parameter: tagQuery")
	}

	if p, err = newSearchPage(mtrApiUrl); err != nil {
		return badRequest("error creating searchPage object")
	}

	if err = p.populateTags(); err != nil {
		return internalServerError(err)
	}

	if err = p.matchingMetrics(tagQuery); err != nil {
		return internalServerError(err)
	}

	if err = p.fetchIcons(); err != nil {
		return internalServerError(err)
	}

	if err := tagsTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}
