package main

import (
	"bytes"
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
	"strings"
)

func tagPut(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	tag := strings.TrimPrefix(r.URL.Path, "/tag/")

	if tag == "" {
		return weft.BadRequest("empty tag")
	}

	if _, err := db.Exec(`INSERT INTO mtr.tag(tag) VALUES($1)`, tag); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			//	no-op.  Nothing to update.
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func tagDelete(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	tag := strings.TrimPrefix(r.URL.Path, "/tag/")

	if tag == "" {
		return weft.BadRequest("empty tag")
	}

	if _, err := db.Exec(`DELETE FROM mtr.tag WHERE tag=$1`, tag); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
