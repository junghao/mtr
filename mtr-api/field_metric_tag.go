package main

import (
	"bytes"
	"github.com/lib/pq"
	"net/http"
)

type fieldMetricTag struct {
	tag
	fieldDevice
	fieldType
}

func (f *fieldMetricTag) loadPK(r *http.Request) *result {
	if res := f.tag.loadPK(r); !res.ok {
		return res
	}

	if res := f.fieldDevice.loadPK(r); !res.ok{
		return res
	}

	if res := f.fieldType.loadPK(r); !res.ok {
		return res
	}

	return &statusOK
}

func (f *fieldMetricTag) save(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"deviceID", "typeID", "tag"}, []string{}); !res.ok {
		return res
	}

	if res := f.loadPK(r); !res.ok {
		return res
	}

	if _, err := db.Exec(`INSERT INTO field.metric_tag(devicePK, typePK, tagPK)
			VALUES($1, $2, $3)`,
		f.devicePK, f.typePK, f.tagPK); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
		} else {
			return internalServerError(err)
		}
	}

	return &statusOK
}

func (f *fieldMetricTag) delete(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"deviceID", "typeID", "tag"}, []string{}); !res.ok {
		return res
	}

	if res := f.loadPK(r); !res.ok {
		return res
	}


	if _, err := db.Exec(`DELETE FROM field.metric_tag
			WHERE devicePK = $1
			AND typePK = $2
			AND tagPK = $3`, f.devicePK, f.typePK, f.tagPK); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}
