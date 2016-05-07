package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/golang/protobuf/proto"
	"github.com/lib/pq"
	"net/http"
	"strings"
	"time"
)

type tag struct {
	tagPK     int
	pk        *result
	tagResult mtrpb.TagSearchResult
}

func (t *tag) save(r *http.Request) *result {
	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	if _, err := db.Exec(`INSERT INTO mtr.tag(tag) VALUES($1)`,
		strings.TrimPrefix(r.URL.Path, "/tag/")); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			//	no-op.  Nothing to update.
		} else {
			return internalServerError(err)
		}
	}

	return &statusOK
}

func (t *tag) delete(r *http.Request) *result {
	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	if _, err := db.Exec(`DELETE FROM mtr.tag WHERE tag=$1`,
		strings.TrimPrefix(r.URL.Path, "/tag/")); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func (t *tag) loadPK(r *http.Request) *result {
	if t.pk == nil {
		tg := r.URL.Query().Get("tag")

		if tg == "" {
			tg = strings.TrimPrefix(r.URL.Path, "/tag/")
			if tg == "" {
				t.pk = badRequest("no tag")
				return t.pk
			}
		}

		if err := dbR.QueryRow(`SELECT tagPK FROM mtr.tag where tag = $1`, tg).Scan(&t.tagPK); err != nil {
			if err == sql.ErrNoRows {
				t.pk = badRequest("unknown tag")
				return t.pk
			}
			t.pk = internalServerError(err)
			return t.pk
		}
		t.pk = &statusOK
	}

	return t.pk
}

func (t *tag) all(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT tag FROM data.latency_tag JOIN mtr.tag USING (tagPK)
				  UNION
				  SELECT tag FROM field.metric_tag JOIN mtr.tag USING (tagPK)
				  ORDER BY TAG ASC
				`); err != nil {
		return internalServerError(err)
	}
	defer rows.Close()

	var ts mtrpb.TagResult

	for rows.Next() {
		var t mtrpb.Tag

		if err = rows.Scan(&t.Tag); err != nil {
			return internalServerError(err)
		}

		ts.Used = append(ts.Used, &t)
	}

	var by []byte
	if by, err = proto.Marshal(&ts); err != nil {
		return internalServerError(err)
	}

	b.Write(by)

	h.Set("Content-Type", "application/x-protobuf")

	return &statusOK
}

func (t *tag) single(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	if res := t.loadPK(r); !res.ok {
		return res
	}

	// Load tagged metrics, latency etc in parallel.
	c1 := t.fieldMetric()
	c2 := t.dataLatency()

	resFinal := &statusOK

	for res := range merge(c1, c2) {
		if !res.ok {
			resFinal = res
		}
	}

	if !resFinal.ok {
		return resFinal
	}

	var by []byte
	var err error
	if by, err = proto.Marshal(&t.tagResult); err != nil {
		return internalServerError(err)
	}

	b.Write(by)

	h.Set("Content-Type", "application/x-protobuf")

	return &statusOK
}

// call loadPK first
func (t *tag) fieldMetric() <-chan *result {
	out := make(chan *result)
	go func() {
		defer close(out)
		var err error
		var rows *sql.Rows

		if rows, err = dbR.Query(`SELECT deviceID, modelID, typeid, time, value, lower, upper
	 			  FROM field.metric_tag JOIN field.metric_summary USING (devicepk, typepk)
			          WHERE tagPK = $1
				`, t.tagPK); err != nil {
			out <- internalServerError(err)
			return
		}
		defer rows.Close()

		var tm time.Time

		for rows.Next() {
			var fmr mtrpb.FieldMetricSummary

			if err = rows.Scan(&fmr.DeviceID, &fmr.ModelID, &fmr.TypeID, &tm, &fmr.Value,
				&fmr.Lower, &fmr.Upper); err != nil {
				out <- internalServerError(err)
				return
			}

			fmr.Seconds = tm.Unix()

			t.tagResult.FieldMetric = append(t.tagResult.FieldMetric, &fmr)
		}

		out <- &statusOK
		return
	}()
	return out
}

// call loadPK first
func (t *tag) dataLatency() <-chan *result {
	out := make(chan *result)
	go func() {
		defer close(out)
		var err error
		var rows *sql.Rows

		if rows, err = dbR.Query(`SELECT siteID, typeID, time, mean, fifty, ninety, lower, upper
	 			  FROM data.latency_tag JOIN data.latency_summary USING (sitePK, typePK)
			          WHERE tagPK = $1
				`, t.tagPK); err != nil {
			out <- internalServerError(err)
			return
		}
		defer rows.Close()

		var tm time.Time

		for rows.Next() {

			var dls mtrpb.DataLatencySummary

			if err = rows.Scan(&dls.SiteID, &dls.TypeID, &tm, &dls.Mean, &dls.Fifty, &dls.Ninety,
				&dls.Lower, &dls.Upper); err != nil {
				out <- internalServerError(err)
				return
			}

			dls.Seconds = tm.Unix()

			t.tagResult.DataLatency = append(t.tagResult.DataLatency, &dls)
		}

		out <- &statusOK
		return
	}()
	return out
}
