package main

import (
	"database/sql"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"strconv"
	"sync"
	"time"
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
var fieldDeviceCache = newCachePK()
var dataSiteCache = newCachePK()


// cache look up types.  These change very rarely.  Additions to the DB
// mean the app server should be restarted.

var fieldTypeCache = struct{
sync.RWMutex
	m map[string]fieldType
}{m: make(map[string]fieldType)}

var dataTypeCache = struct{
	sync.RWMutex
	m map[string]dataType
}{m: make(map[string]dataType)}

// Domain types - these reflect database tables.
// The tables have a primary key xxxPK (pk) and an identifier xxxID (id).
// There are read methods to load the primary keys based on the id.
// this is very good for DB performance.  It also allows in memory PK caching
// e.g., for the application PKs.
// It does introduce race conditions from the point of view of the clients
// sending data to mtr-api.  The race being the time between calling the read
// methods and saving metrics etc.  These should not be significant and as long as
// clients handle 500s and resend should never result in data loss.

// refer to database/ddl/*.ddl for definitions of tables etc.

// tag - table mtr.tag
// tags can be applied metrics, latencies etc.
type tag struct {
	pk int
	id string
}

// tagSearch for tag search results.
// needed for use with singleProto and fan out.
type tagSearch struct {
	tag
	tagResult mtrpb.TagSearchResult
}

// fieldModel - table field.model
// field devices have a model.
type fieldModel struct {
	pk int
	id string
}

// fieldDevice table field.device
// a device e.g., a seismic data logger that is located at a point.
type fieldDevice struct {
	pk                  int
	id                  string
	longitude, latitude float64
	fieldModel          // only used during create.  Not bothering loading pk in read().
}

// fieldType - table field.type
type fieldType struct {
	id    string
	pk    int
	scale float64 // used to scale the stored metric for display
	display  string // display unit after the metric has been multiplied by scale.
}

// fieldDeviceType field metrics are for a type from a device.
// one device can have several metrics.
// see devicePK and typePK in table - field.metric and the PK on that table.
type fieldDeviceType struct {
	fieldDevice
	fieldType
}

// fieldMetric - table field.metric
// metrics have a value at a time.
type fieldMetric struct {
	fieldDeviceType
	val int
	t   time.Time
}

// fieldMetricTag - table field.metric_tag
type fieldMetricTag struct {
	fieldDeviceType
	tag
}

// fieldThreshold - table field.threshold
// to be considered good metrics must be within the thresholds.
type fieldThreshold struct {
	fieldDeviceType
	lower, upper int
}

// fieldLatest - for get queries.
type fieldLatest struct {
	typeID string
}

// dataSite - table data.site
// data is recorded at a site which is located at a point.
type dataSite struct {
	pk                  int
	id                  string
	longitude, latitude float64
}

// dataType - table data.type
type dataType struct {
	id    string
	pk    int
	scale float64 // used to scale the stored metric for display
	display  string // display unit after the metric has been multiplied by scale.
}

// dataSiteType data metrics are a type from a site.
// A single site can produce many types of metrics.
// see sitePK and typePK in table data.latency
type dataSiteType struct {
	dataSite
	dataType
}

// dataLatency - table data.latency
type dataLatency struct {
	dataSiteType
	t                             time.Time
	mean, min, max, fifty, ninety int
}

// dataLatencyThreshold - table data.latency_threshold
type dataLatencyThreshold struct {
	dataSiteType
	lower, upper int
}

// dataLatencyTag - table data.latency_tag
type dataLatencyTag struct {
	dataSiteType
	tag
}

// for SVG maps.
type point struct {
	latitude, longitude float64
	x, y                float64
}

// application - table app.application
type application struct {
	pk int
	id string // usually the application name.
}

// applicationInstance - table app.instance
// there might be multiple instances of an app running.
type applicationInstance struct {
	pk int
	id string // usually the host name.
}

// applicationSource - table app.source
type applicationSource struct {
	pk int
	id string // usually the name of something being timed e.g., a function name.
}

// applicationType - table app.type
// super set of HTTP status codes.
// see also internal/mtr_const.go
type applicationType struct {
	pk int
	id string
}

// applicationCounter - table app.counter
// things like HTTP requests, messages sent etc.
type applicationCounter struct {
	application
	applicationType
	t time.Time
	c int
}

// appMetric - table app.metric
// things like memory, routines, object count.
type applicationMetric struct {
	application
	applicationInstance
	applicationType
	t     time.Time
	value int64
}

// applicationTimer app.timer
// for timing things.
type applicationTimer struct {
	application
	applicationSource
	t                             time.Time
	average, count, fifty, ninety int
}

//appMetric for get requests.
type appMetric struct {
	application
}

// InstanceMetric for sorting instances for SVG plots.
// public for use with sort.
type InstanceMetric struct {
	instancePK, typePK int
}

type InstanceMetrics []InstanceMetric

func (l InstanceMetrics) Len() int           { return len(l) }
func (l InstanceMetrics) Less(i, j int) bool { return l[i].instancePK < l[j].instancePK }
func (l InstanceMetrics) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

// read methods set the relevant id members and then load the pk members.

func (a *dataLatencyTag) read(siteID, typeID, tag string) *weft.Result {
	if res := a.dataSiteType.read(siteID, typeID); !res.Ok {
		return res
	}

	if res := a.tag.read(tag); !res.Ok {
		return res
	}

	return &weft.StatusOK
}

func (a *fieldMetricTag) read(deviceID, typeID, tag string) *weft.Result {
	if res := a.fieldDeviceType.read(deviceID, typeID); !res.Ok {
		return res
	}

	if res := a.tag.read(tag); !res.Ok {
		return res
	}

	return &weft.StatusOK
}

func (a *tag) read(tag string) *weft.Result {
	if tag == "" {
		return weft.BadRequest("empty tag")
	}

	a.id = tag

	if err := dbR.QueryRow(`SELECT tagPK FROM mtr.tag where tag = $1`, a.id).Scan(&a.pk); err != nil {
		if err == sql.ErrNoRows {
			return weft.BadRequest("tag not found")
		}
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (a *dataType) read(typeID string) *weft.Result {
	if typeID == "" {
		return weft.BadRequest("empty typeID")
	}

	a.id = typeID

	var ok bool
	var b dataType

	dataTypeCache.RLock()
	b, ok = dataTypeCache.m[a.id]
	dataTypeCache.RUnlock()

	if ok {
		a.pk = b.pk
		a.display = b.display
		a.scale = b.scale
		return &weft.StatusOK
	}

	dataTypeCache.Lock()
	defer dataTypeCache.Unlock()

	if err := dbR.QueryRow(`SELECT typePK, scale, display FROM data.type where typeID = $1`,
		a.id).Scan(&a.pk, &a.scale, &a.display); err != nil {
		if err == sql.ErrNoRows {
			return &weft.NotFound
		}
		return weft.InternalServerError(err)
	}

	dataTypeCache.m[a.id] = *a

	return &weft.StatusOK
}

func (a *dataSiteType) read(siteID, typeID string) *weft.Result {
	if res := a.dataType.read(typeID); !res.Ok {
		return res
	}

	if res := a.dataSite.read(siteID); !res.Ok {
		return res
	}

	return &weft.StatusOK
}

func (a *dataSite) read(siteID string) *weft.Result {
	if siteID == "" {
		return weft.BadRequest("empty siteID")
	}

	a.id = siteID

	var ok bool

	dataSiteCache.RLock()
	a.pk, ok = dataSiteCache.m[a.id]
	dataSiteCache.RUnlock()

	if ok {
		return &weft.StatusOK
	}

	dataSiteCache.Lock()
	defer dataSiteCache.Unlock()


	if err := dbR.QueryRow(`SELECT sitePK FROM data.site where siteID = $1`,
		a.id).Scan(&a.pk); err != nil {
		if err == sql.ErrNoRows {
			return weft.BadRequest("unknown siteID")
		}
		return weft.InternalServerError(err)
	}

	dataSiteCache.m[a.id] = a.pk

	return &weft.StatusOK
}

func (a *fieldType) read(typeID string) *weft.Result {
	if typeID == "" {
		return weft.BadRequest("empty typeID")
	}

	a.id = typeID

	var ok bool
	var b fieldType

	fieldTypeCache.RLock()
	b, ok = fieldTypeCache.m[a.id]
	fieldTypeCache.RUnlock()

	if ok {
		a.pk = b.pk
		a.display = b.display
		a.scale = b.scale
		return &weft.StatusOK
	}

	fieldTypeCache.Lock()
	defer fieldTypeCache.Unlock()

	if err := dbR.QueryRow(`SELECT typePK, scale, display FROM field.type where typeID = $1`,
		a.id).Scan(&a.pk, &a.scale, &a.display); err != nil {
		if err == sql.ErrNoRows {
			return &weft.NotFound
		}
		return weft.InternalServerError(err)
	}

	fieldTypeCache.m[a.id] = *a

	return &weft.StatusOK
}


func (a *fieldDevice) read(deviceID string) *weft.Result {
	if deviceID == "" {
		return weft.BadRequest("empty deviceID")
	}

	a.id = deviceID

	var ok bool

	fieldDeviceCache.RLock()
	a.pk, ok = fieldDeviceCache.m[a.id]
	fieldDeviceCache.RUnlock()

	if ok {
		return &weft.StatusOK
	}

	fieldDeviceCache.Lock()
	defer fieldDeviceCache.Unlock()

	if err := dbR.QueryRow(`SELECT devicePK FROM field.device where deviceID = $1`,
		a.id).Scan(&a.pk); err != nil {
		if err == sql.ErrNoRows {
			return weft.BadRequest("unknown deviceID")
		}
		return weft.InternalServerError(err)
	}

	fieldDeviceCache.m[a.id] = a.pk

	return &weft.StatusOK
}

func (a *fieldDeviceType) read(deviceID, typeID string) *weft.Result {
	if res := a.fieldDevice.read(deviceID); !res.Ok {
		return res
	}

	if res := a.fieldType.read(typeID); !res.Ok {
		return res
	}

	return &weft.StatusOK
}

func (a *fieldModel) read(modelID string) *weft.Result {
	if modelID == "" {
		return weft.BadRequest("empty modelID")
	}

	a.id = modelID

	if err := dbR.QueryRow(`SELECT modelPK FROM field.model where modelID = $1`, a.id).Scan(&a.pk); err != nil {
		if err == sql.ErrNoRows {
			return weft.BadRequest("unknown modelID")
		}
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (a *applicationType) read(applicationTypeID string) *weft.Result {
	if applicationTypeID == "" {
		return weft.BadRequest("empty applicationTypeID")
	}

	a.id = applicationTypeID

	// TODO could validate this without hitting the DB
	// use a hash in internal?
	var err error
	if a.pk, err = strconv.Atoi(a.id); err != nil {
		return weft.BadRequest("invalid typeID")
	}

	return &weft.StatusOK
}

// readCreate methods for application metrics.  These are more complicated
// then read() methods.  When application metrics are sent the application,
// instance, and source may not yet exist in the DB so they may need to be
// created.
//
// Memory caching is also used.

func (a *application) readCreate(applicationID string) *weft.Result {
	if applicationID == "" {
		return weft.BadRequest("empty applicationID")
	}

	a.id = applicationID

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


func (a *applicationInstance) readCreate(instanceID string) *weft.Result {
	if instanceID == "" {
		return weft.BadRequest("empty instanceID")
	}

	a.id = instanceID

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

func (a *applicationSource) readCreate(sourceID string) *weft.Result {
	if sourceID == "" {
		return weft.BadRequest("empty sourceID")
	}

	a.id = sourceID

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

// del deletes all metrics from for the application from the db.
// used with testing.
func (a *application) del(applicationID string) *weft.Result {
	if applicationID == "" {
		return weft.BadRequest("empty applicationID")
	}

	a.id = applicationID

	applicationCache.Lock()
	defer applicationCache.Unlock()

	if _, err := db.Exec(`DELETE FROM app.application WHERE applicationID = $1`,
		a.id); err != nil {
		return weft.InternalServerError(err)
	}

	delete(applicationCache.m, a.id)

	return &weft.StatusOK
}
