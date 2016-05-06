package main

import (
	"database/sql"
	"github.com/lib/pq"
	"net/http"
)

type tag struct {
	tagPK int
}

func (t *tag) save(r *http.Request) *result {
	if res := checkQuery(r, []string{"tag"}, []string{}); !res.ok {
		return res
	}

	if _, err := db.Exec(`INSERT INTO mtr.tag(tag) VALUES($1)`,
		r.URL.Query().Get("tag")); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			//	no-op.  Nothing to update.
		} else {
			return internalServerError(err)
		}
	}

	return &statusOK
}

func (t *tag) delete(r *http.Request) *result {
	if res := checkQuery(r, []string{"tag"}, []string{}); !res.ok {
		return res
	}

	if _, err := db.Exec(`DELETE FROM mtr.tag WHERE tag=$1`,
		r.URL.Query().Get("tag")); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func (t *tag) loadPK(r *http.Request) *result {
	if err := dbR.QueryRow(`SELECT tagPK FROM mtr.tag where tag = $1`,
		r.URL.Query().Get("tag")).Scan(&t.tagPK); err != nil {
		if err == sql.ErrNoRows {
			return badRequest("unknown tag")
		}
		return internalServerError(err)
	}

	return &statusOK
}
