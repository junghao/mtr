package main

import (
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
	"strings"
)

// tag - table mtr.tag
// tags can be applied metrics, latencies etc.
type tag struct{}

func (a tag) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

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

func (a tag) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	tag := strings.TrimPrefix(r.URL.Path, "/tag/")

	if tag == "" {
		return weft.BadRequest("empty tag")
	}

	if _, err := db.Exec(`DELETE FROM mtr.tag WHERE tag=$1`, tag); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
