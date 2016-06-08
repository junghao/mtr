package main

import (
	"bytes"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"net/http"
)

func appPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error

	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	// Parse mtr-api /app endpoint as a protobuf
	u := *mtrApiUrl
	u.Path = "/app"

	var apps []byte
	if apps, err = getBytes(u.String(), "application/x-protobuf"); err != nil {
		return weft.InternalServerError(err)
	}

	var appsPb mtrpb.AppIDSummaryResult

	if err = proto.Unmarshal(apps, &appsPb); err != nil {
		return weft.InternalServerError(err)
	}

	// We create a page struct with variables to substitute into the loaded template
	p := mtrUiPage{}
	p.Path = r.URL.Path
	p.Border.Title = "GeoNet MTR - applications"

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	for _, appId := range appsPb.Result {
		p.AppIDs = append(p.AppIDs, app{ID: appId.ApplicationID})
	}

	if err = appTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

// Contruct a simple page that has several plots (and time intervals as param) for the specified applicationID param
func appPlotPageHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error

	if res := weft.CheckQuery(r, []string{"applicationID"}, []string{"resolution"}); !res.Ok {
		return res
	}

	p := mtrUiPage{}
	p.Path = r.URL.Path
	p.MtrApiUrl = mtrApiUrl.String()
	p.Border.Title = "GeoNet MTR - application ID"

	// get the applicationID and resolution params from the URL
	p.pageParam(r.URL.Query())

	if p.Resolution == "" {
		p.Resolution = "minute"
	}

	if err = p.populateTags(); err != nil {
		return weft.InternalServerError(err)
	}

	if err = appPlotTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
