package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"github.com/lib/pq"
	"net/http"
	"strings"
	"time"
)

type tag struct {
	tagPK     int
	pk        *weft.Result
	tagResult mtrpb.TagSearchResult
}

func (t *tag) save(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	if _, err := db.Exec(`INSERT INTO mtr.tag(tag) VALUES($1)`,
		strings.TrimPrefix(r.URL.Path, "/tag/")); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			//	no-op.  Nothing to update.
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func (t *tag) delete(r *http.Request) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	if _, err := db.Exec(`DELETE FROM mtr.tag WHERE tag=$1`,
		strings.TrimPrefix(r.URL.Path, "/tag/")); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (t *tag) loadPK(r *http.Request) *weft.Result {
	if t.pk == nil {
		tg := r.URL.Query().Get("tag")

		if tg == "" {
			tg = strings.TrimPrefix(r.URL.Path, "/tag/")
			if tg == "" {
				t.pk = weft.BadRequest("no tag")
				return t.pk
			}
		}

		if err := dbR.QueryRow(`SELECT tagPK FROM mtr.tag where tag = $1`, tg).Scan(&t.tagPK); err != nil {
			if err == sql.ErrNoRows {
				t.pk = weft.BadRequest("unknown tag")
				return t.pk
			}
			t.pk = weft.InternalServerError(err)
			return t.pk
		}
		t.pk = &weft.StatusOK
	}

	return t.pk
}

func (t *tag) all(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT tag FROM data.latency_tag JOIN mtr.tag USING (tagPK)
				  UNION
				  SELECT tag FROM field.metric_tag JOIN mtr.tag USING (tagPK)
				  ORDER BY TAG ASC
				`); err != nil {
		return weft.InternalServerError(err)
	}
	defer rows.Close()

	var ts mtrpb.TagResult

	for rows.Next() {
		var t mtrpb.Tag

		if err = rows.Scan(&t.Tag); err != nil {
			return weft.InternalServerError(err)
		}

		ts.Used = append(ts.Used, &t)
	}

	var by []byte
	if by, err = proto.Marshal(&ts); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	h.Set("Content-Type", "application/x-protobuf")

	return &weft.StatusOK
}

func (t *tag) single(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	if res := t.loadPK(r); !res.Ok {
		return res
	}

	// Load tagged metrics, latency etc in parallel.
	c1 := t.fieldMetric()
	c2 := t.dataLatency()

	resFinal := &weft.StatusOK

	for res := range merge(c1, c2) {
		if !res.Ok {
			resFinal = res
		}
	}

	if !resFinal.Ok {
		return resFinal
	}

	var by []byte
	var err error
	if by, err = proto.Marshal(&t.tagResult); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	h.Set("Content-Type", "application/x-protobuf")

	return &weft.StatusOK
}

// call loadPK first
func (t *tag) fieldMetric() <-chan *weft.Result {
	out := make(chan *weft.Result)
	go func() {
		defer close(out)
		var err error
		var rows *sql.Rows

		if rows, err = dbR.Query(`SELECT deviceID, modelID, typeid, time, value, lower, upper
	 			  FROM field.metric_tag JOIN field.metric_summary USING (devicepk, typepk)
			          WHERE tagPK = $1
				`, t.tagPK); err != nil {
			out <- weft.InternalServerError(err)
			return
		}
		defer rows.Close()

		var tm time.Time

		for rows.Next() {
			var fmr mtrpb.FieldMetricSummary

			if err = rows.Scan(&fmr.DeviceID, &fmr.ModelID, &fmr.TypeID, &tm, &fmr.Value,
				&fmr.Lower, &fmr.Upper); err != nil {
				out <- weft.InternalServerError(err)
				return
			}

			fmr.Seconds = tm.Unix()

			t.tagResult.FieldMetric = append(t.tagResult.FieldMetric, &fmr)
		}

		out <- &weft.StatusOK
		return
	}()
	return out
}

// call loadPK first
func (t *tag) dataLatency() <-chan *weft.Result {
	out := make(chan *weft.Result)
	go func() {
		defer close(out)
		var err error
		var rows *sql.Rows

		if rows, err = dbR.Query(`SELECT siteID, typeID, time, mean, fifty, ninety, lower, upper
	 			  FROM data.latency_tag JOIN data.latency_summary USING (sitePK, typePK)
			          WHERE tagPK = $1
				`, t.tagPK); err != nil {
			out <- weft.InternalServerError(err)
			return
		}
		defer rows.Close()

		var tm time.Time

		for rows.Next() {

			var dls mtrpb.DataLatencySummary

			if err = rows.Scan(&dls.SiteID, &dls.TypeID, &tm, &dls.Mean, &dls.Fifty, &dls.Ninety,
				&dls.Lower, &dls.Upper); err != nil {
				out <- weft.InternalServerError(err)
				return
			}

			dls.Seconds = tm.Unix()

			t.tagResult.DataLatency = append(t.tagResult.DataLatency, &dls)
		}

		out <- &weft.StatusOK
		return
	}()
	return out
}
