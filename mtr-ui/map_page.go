package main

import (
	"bytes"
	"fmt"
	"github.com/GeoNet/weft"
	"net/http"
	"strings"
)

type mapPage struct {
	page
	ActiveTab string
	MtrApiUrl string
	TypeID    string
	MapApiUrl string
}

func mapPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {

	var err error

	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	p := mapPage{}
	p.MtrApiUrl = mtrApiUrl.String()
	p.Border.Title = "GeoNet MTR - Map"
	p.ActiveTab = "Map"

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	if err = p.populateTypes(); err != nil {
		return weft.InternalServerError(err)
	}

	s := strings.TrimPrefix(r.URL.Path, "/map")

	typeExist := false
	if ArrayContains(s, []string{"", "/"}) {
		p.TypeID = ""
		typeExist = true
		for _, mapdef := range p.Border.MapList {
			if len(mapdef.TypeIDs) > 0 {
				p.TypeID = mapdef.TypeIDs[0]
				p.MapApiUrl = mapdef.ApiUrl
				break
			}
		}
	} else {
		s1 := strings.TrimPrefix(s, "/")
		for _, mapdef := range p.Border.MapList {
			if ArrayContains(s1, mapdef.TypeIDs) {
				p.TypeID = s1
				p.MapApiUrl = mapdef.ApiUrl
				typeExist = true
				break
			}
		}
	}

	if !typeExist {
		return weft.InternalServerError(fmt.Errorf("Unknown map type"))
	}

	if err = mapTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func ArrayContains(v string, a []string) bool {
	for _, i := range a {
		if i == v {
			return true
		}
	}
	return false
}
