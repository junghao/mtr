package main

import (
	"database/sql"
	"fmt"
	"github.com/GeoNet/weft"
	"strconv"
	"sync"
)

// cachePK is for caching primary keys that do not change very often
// and so are very read heavy from the db.  Using this makes this
// app unsuitable for horizontal scaling with http.DELETE methods.
// Other approaches would be to use groupcache or a read only DB instance.
type cachePK struct {
	sync.RWMutex
	m map[string]int
}

func newCachePK() cachePK {
	return cachePK{m: make(map[string]int)}
}

var applicationCache = newCachePK()
var applicationInstanceCache = newCachePK()
var applicationSourceCache = newCachePK()

type application struct {
	pk int
	id string
}

type applicationInstance struct {
	pk int
	id string
}

type applicationSource struct {
	pk int
	id string
}

type applicationType struct {
	pk int
	id string
}

// create sets pk to the db primary key for id,
// creating it in the db if needed.
func (a *application) create() *weft.Result {
	if a.id == "" {
		return weft.InternalServerError(fmt.Errorf("empty application.id"))
	}

	var ok bool

	applicationCache.RLock()
	a.pk, ok = applicationCache.m[a.id]
	applicationCache.RUnlock()

	if ok {
		return &weft.StatusOK
	}

	applicationCache.Lock()
	defer applicationCache.Unlock()

	err := dbR.QueryRow(`SELECT applicationPK FROM app.application WHERE applicationID = $1`,
		a.id).Scan(&a.pk)
	switch err {
	case nil:
	case sql.ErrNoRows:
		if _, err = db.Exec(`INSERT INTO app.application(applicationID) VALUES($1)`, a.id); err != nil {
			return weft.InternalServerError(err)
		}
		if err = db.QueryRow(`SELECT applicationPK FROM app.application WHERE applicationID = $1`,
			a.id).Scan(&a.pk); err != nil {
			return weft.InternalServerError(err)
		}
	default:
		return weft.InternalServerError(err)
	}

	applicationCache.m[a.id] = a.pk

	return &weft.StatusOK
}

// read sets pk to the database primary key.
func (a *application) read() *weft.Result {
	if a.id == "" {
		return weft.InternalServerError(fmt.Errorf("empty application.id"))
	}

	var ok bool

	applicationCache.RLock()
	a.pk, ok = applicationCache.m[a.id]
	applicationCache.RUnlock()

	if ok {
		return &weft.StatusOK
	}

	err := dbR.QueryRow(`SELECT applicationPK FROM app.application WHERE applicationID = $1`,
		a.id).Scan(&a.pk)
	switch err {
	case nil:
		applicationCache.Lock()
		applicationCache.m[a.id] = a.pk
		applicationCache.Unlock()

		return &weft.StatusOK
	case sql.ErrNoRows:
		return &weft.NotFound
	default:
		return weft.InternalServerError(err)
	}
}

// del deletes all metrics from for the application from the db.
func (a *application) del() *weft.Result {
	if a.id == "" {
		return weft.InternalServerError(fmt.Errorf("empty application.id"))
	}

	applicationCache.Lock()
	defer applicationCache.Unlock()

	if _, err := db.Exec(`DELETE FROM app.application WHERE applicationID = $1`,
		a.id); err != nil {
		return weft.InternalServerError(err)
	}

	delete(applicationCache.m, a.id)

	return &weft.StatusOK
}

// create sets pk to the db primary key for id,
// creating it in the db if needed.
func (a *applicationInstance) create() *weft.Result {
	if a.id == "" {
		return weft.InternalServerError(fmt.Errorf("empty applicationInstance.id"))
	}

	var ok bool

	applicationInstanceCache.RLock()
	a.pk, ok = applicationInstanceCache.m[a.id]
	applicationInstanceCache.RUnlock()

	if ok {
		return &weft.StatusOK
	}

	applicationInstanceCache.Lock()
	defer applicationInstanceCache.Unlock()

	err := dbR.QueryRow(`SELECT instancePK FROM app.instance WHERE instanceID = $1`,
		a.id).Scan(&a.pk)
	switch err {
	case nil:
	case sql.ErrNoRows:
		if _, err = db.Exec(`INSERT INTO app.instance(instanceID) VALUES($1)`,
			a.id); err != nil {
			return weft.InternalServerError(err)
		}
		if err = db.QueryRow(`SELECT instancePK FROM app.instance WHERE instanceID = $1`,
			a.id).Scan(&a.pk); err != nil {
			return weft.InternalServerError(err)
		}
	default:
		return weft.InternalServerError(err)
	}

	applicationInstanceCache.m[a.id] = a.pk

	return &weft.StatusOK
}

// create sets pk to the db primary key for id,
// creating it in the db if needed.
func (a *applicationSource) create() *weft.Result {
	if a.id == "" {
		return weft.InternalServerError(fmt.Errorf("empty applicationSource.id"))
	}

	var ok bool

	applicationSourceCache.RLock()
	a.pk, ok = applicationSourceCache.m[a.id]
	applicationSourceCache.RUnlock()

	if ok {
		return &weft.StatusOK
	}

	applicationSourceCache.Lock()
	defer applicationSourceCache.Unlock()

	err := dbR.QueryRow(`SELECT sourcePK FROM app.source WHERE sourceID = $1`,
		a.id).Scan(&a.pk)

	switch err {
	case nil:
	case sql.ErrNoRows:
		if _, err = db.Exec(`INSERT INTO app.source(sourceID) VALUES($1)`,
			a.id); err != nil {
			return weft.InternalServerError(err)
		}
		if err = db.QueryRow(`SELECT sourcePK FROM app.source WHERE sourceID = $1`,
			a.id).Scan(&a.pk); err != nil {
			return weft.InternalServerError(err)
		}
	default:
		return weft.InternalServerError(err)
	}

	applicationSourceCache.m[a.id] = a.pk

	return &weft.StatusOK
}

// read sets pk to the database primary key.
func (a *applicationType) read() *weft.Result {
	if a.id == "" {
		return weft.InternalServerError(fmt.Errorf("empty applicationType.id"))
	}

	// TODO could validate this without hitting the DB
	var err error
	if a.pk, err = strconv.Atoi(a.id); err != nil {
		return weft.BadRequest("invalid typeID")
	}

	return &weft.StatusOK
}
