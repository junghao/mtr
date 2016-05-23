package main

import (
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
	"strings"
)

func (a *tag) put(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	a.id = strings.TrimPrefix(r.URL.Path, "/tag/")

	if a.id == "" {
		return weft.BadRequest("empty tag")
	}

	if _, err := db.Exec(`INSERT INTO mtr.tag(tag) VALUES($1)`, a.id); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			//	no-op.  Nothing to update.
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func (a *tag) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	a.id = strings.TrimPrefix(r.URL.Path, "/tag/")

	if a.id == "" {
		return weft.BadRequest("empty tag")
	}

	if _, err := db.Exec(`DELETE FROM mtr.tag WHERE tag=$1`, a.id); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
