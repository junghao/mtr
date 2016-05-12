package main

import (
	"bytes"
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
)

type fieldMetricTag struct {
	tag
	fieldDevice
	fieldType
}

func (f *fieldMetricTag) loadPK(r *http.Request) *weft.Result {
	if res := f.tag.loadPK(r); !res.Ok {
		return res
	}

	if res := f.fieldDevice.loadPK(r); !res.Ok {
		return res
	}

	if res := f.fieldType.loadPK(r); !res.Ok {
		return res
	}

	return &weft.StatusOK
}

func (f *fieldMetricTag) save(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "typeID", "tag"}, []string{}); !res.Ok {
		return res
	}

	if res := f.loadPK(r); !res.Ok {
		return res
	}

	if _, err := db.Exec(`INSERT INTO field.metric_tag(devicePK, typePK, tagPK)
			VALUES($1, $2, $3)`,
		f.devicePK, f.typePK, f.tagPK); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func (f *fieldMetricTag) delete(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"deviceID", "typeID", "tag"}, []string{}); !res.Ok {
		return res
	}

	if res := f.loadPK(r); !res.Ok {
		return res
	}

	if _, err := db.Exec(`DELETE FROM field.metric_tag
			WHERE devicePK = $1
			AND typePK = $2
			AND tagPK = $3`, f.devicePK, f.typePK, f.tagPK); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
