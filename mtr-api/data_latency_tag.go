package main

import (
	"bytes"
	"github.com/lib/pq"
	"net/http"
)

type dataLatencyTag struct {
	tag
	dataSite
	dataType
}

func (f *dataLatencyTag) loadPK(r *http.Request) *result {
	if res := f.tag.loadPK(r); !res.ok {
		return res
	}

	if res := f.dataType.load(r); !res.ok {
		return res
	}

	if res := f.dataSite.loadPK(r); !res.ok {
		return res
	}

	return &statusOK
}

func (f *dataLatencyTag) save(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"siteID", "typeID", "tag"}, []string{}); !res.ok {
		return res
	}

	if res := f.loadPK(r); !res.ok {
		return res
	}

	if _, err := db.Exec(`INSERT INTO data.latency_tag(sitePK, typePK, tagPK)
			VALUES($1, $2, $3)`,
		f.sitePK, f.typePK, f.tagPK); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
		} else {
			return internalServerError(err)
		}
	}

	return &statusOK
}

func (f *dataLatencyTag) delete(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{"siteID", "typeID", "tag"}, []string{}); !res.ok {
		return res
	}

	if res := f.loadPK(r); !res.ok {
		return res
	}

	if _, err := db.Exec(`DELETE FROM data.latency_tag
			WHERE sitePK = $1
			AND typePK = $2
			AND tagPK = $3`, f.sitePK, f.typePK, f.tagPK); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}
