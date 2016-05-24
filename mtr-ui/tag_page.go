package main

import (
	"bytes"
	"fmt"
	"github.com/GeoNet/weft"
	"net/http"
	"strings"
)

type tagPage struct {
	page
	Path    string
	TagTabs []string
	Tags    []string
}

var tagGrouper = []string{"ABC", "DEF", "GHI", "JKL", "MNO", "POR", "STU", "VWXYZ", "0123456789"}

func tagPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error

	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	p := tagPage{}
	p.Path = r.URL.Path
	p.Border.Title = "GeoNet MTR"

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	ph := strings.TrimPrefix(p.Path, "/tag")
	if strings.HasPrefix(ph, "/") {
		ph = ph[1:]
	}

	currTab := -1

	// Create grouping tabs
	p.TagTabs = make([]string, 0)
	for i, k := range tagGrouper {
		s := k[:1] + "-" + k[len(k)-1:]
		p.TagTabs = append(p.TagTabs, s)
		if ph == s {
			currTab = i
		}
	}

	if currTab == -1 && ph != "" {
		return weft.BadRequest("Invalid tag index.")
	}

	if ph == "" {
		currTab = 0
	}

	p.Path = p.TagTabs[currTab]

	for _, t := range p.Border.TagList {
		c := t[:1]
		if strings.Contains(tagGrouper[currTab], c) {
			p.Tags = append(p.Tags, t)
		}
	}

	if err = tagPageTemplate.ExecuteTemplate(b, "border", p); err != nil {
		fmt.Println(err)
		return weft.InternalServerError(err)
	}
	return &weft.StatusOK
}
