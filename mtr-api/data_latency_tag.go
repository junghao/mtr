package main

import (
	"bytes"
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
)

type dataLatencyTag struct {
	tag
	dataSite
	dataType
}

func (f *dataLatencyTag) loadPK(r *http.Request) *weft.Result {
	if res := f.tag.loadPK(r); !res.Ok {
		return res
	}

	if res := f.dataType.load(r); !res.Ok {
		return res
	}

	if res := f.dataSite.loadPK(r); !res.Ok {
		return res
	}

	return &weft.StatusOK
}

func (f *dataLatencyTag) save(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID", "tag"}, []string{}); !res.Ok {
		return res
	}

	if res := f.loadPK(r); !res.Ok {
		return res
	}

	if _, err := db.Exec(`INSERT INTO data.latency_tag(sitePK, typePK, tagPK)
			VALUES($1, $2, $3)`,
		f.sitePK, f.typePK, f.tagPK); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func (f *dataLatencyTag) delete(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID", "tag"}, []string{}); !res.Ok {
		return res
	}

	if res := f.loadPK(r); !res.Ok {
		return res
	}

	if _, err := db.Exec(`DELETE FROM data.latency_tag
			WHERE sitePK = $1
			AND typePK = $2
			AND tagPK = $3`, f.sitePK, f.typePK, f.tagPK); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
