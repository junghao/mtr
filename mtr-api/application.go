package main

import (
	"database/sql"
	"github.com/GeoNet/weft"
	"net/http"
	"strconv"
)

type application struct {
	applicationPK int
	applicationID string
}

type applicationInstance struct {
	instancePK int
	instanceID string
}

type applicationSource struct {
	sourcePK int
	sourceID string
}

type applicationType struct {
	typePK int
}

// Find  (and possibly create) the applicationPK for the applicationID
func (a *application) loadPK(r *http.Request) *weft.Result {
	a.applicationID = r.URL.Query().Get("applicationID")

	err := db.QueryRow(`SELECT applicationPK FROM app.application WHERE applicationID = $1`,
		a.applicationID).Scan(&a.applicationPK)
	switch err {
	case nil:
	case sql.ErrNoRows:
		if _, err = db.Exec(`INSERT INTO app.application(applicationID) VALUES($1)`, a.applicationID); err != nil {
			// TODO ignoring error due to race on insert between calls to this func.  Use a transaction here?
			//return weft.InternalServerError(err)
		}
		if err = db.QueryRow(`SELECT applicationPK FROM app.application WHERE applicationID = $1`, a.applicationID).Scan(&a.applicationPK); err != nil {
			return weft.InternalServerError(err)
		}
	default:
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (a *application) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{"applicationID"}, []string{}); !res.Ok {
		return res
	}

	if res := a.loadPK(r); !res.Ok {
		return res
	}

	if _, err := db.Exec(`DELETE FROM app.application WHERE applicationID = $1`,
		a.applicationID); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

// Find  (and possibly create) the instancePK for the instanceID
func (i *applicationInstance) loadPK(r *http.Request) *weft.Result {
	i.instanceID = r.URL.Query().Get("instanceID")

	err := db.QueryRow(`SELECT instancePK FROM app.instance WHERE instanceID = $1`,
		i.instanceID).Scan(&i.instancePK)
	switch err {
	case nil:
	case sql.ErrNoRows:
		if _, err = db.Exec(`INSERT INTO app.instance(instanceID) VALUES($1)`,
			i.instanceID); err != nil {
			// TODO ignoring error due to race on insert between calls to this func.  Use a transaction here?
			//return weft.InternalServerError(err)
		}
		if err = db.QueryRow(`SELECT instancePK FROM app.instance WHERE instanceID = $1`,
			i.instanceID).Scan(&i.instancePK); err != nil {
			return weft.InternalServerError(err)
		}
	default:
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

// Find  (and possibly create) the sourcePK for the sourceID
func (s *applicationSource) loadPK(r *http.Request) *weft.Result {
	s.sourceID = r.URL.Query().Get("sourceID")

	err := db.QueryRow(`SELECT sourcePK FROM app.source WHERE sourceID = $1`,
		s.sourceID).Scan(&s.sourcePK)

	switch err {
	case nil:
	case sql.ErrNoRows:
		if _, err = db.Exec(`INSERT INTO app.source(sourceID) VALUES($1)`,
			s.sourceID); err != nil {
			// TODO ignoring error due to race on insert between calls to this func.  Use a transaction here?
		}
		if err = db.QueryRow(`SELECT sourcePK FROM app.source WHERE sourceID = $1`,
			s.sourceID).Scan(&s.sourcePK); err != nil {
			return weft.InternalServerError(err)
		}
	default:
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (a *applicationType) loadPK(r *http.Request) *weft.Result {
	// TODO could validate this without hitting the DB
	var err error
	if a.typePK, err = strconv.Atoi(r.URL.Query().Get("typeID")); err != nil {
		return weft.BadRequest("invalid typeID")
	}

	return &weft.StatusOK
}
