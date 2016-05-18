package main

import (
	"bytes"
	"github.com/GeoNet/weft"
	"github.com/lib/pq"
	"net/http"
	"database/sql"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/golang/protobuf/proto"
)

type dataLatencyTag struct {
	tag
	dataSite
	dataType
}

func (f *dataLatencyTag) loadPK(r *http.Request) *weft.Result {
	if res := f.tag.loadPK(r); !res.Ok {
		return res
	}

	if res := f.dataType.load(r); !res.Ok {
		return res
	}

	if res := f.dataSite.loadPK(r); !res.Ok {
		return res
	}

	return &weft.StatusOK
}

func (f *dataLatencyTag) save(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID", "tag"}, []string{}); !res.Ok {
		return res
	}

	if res := f.loadPK(r); !res.Ok {
		return res
	}

	if _, err := db.Exec(`INSERT INTO data.latency_tag(sitePK, typePK, tagPK)
			VALUES($1, $2, $3)`,
		f.sitePK, f.typePK, f.tagPK); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func (f *dataLatencyTag) delete(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID", "tag"}, []string{}); !res.Ok {
		return res
	}

	if res := f.loadPK(r); !res.Ok {
		return res
	}

	if _, err := db.Exec(`DELETE FROM data.latency_tag
			WHERE sitePK = $1
			AND typePK = $2
			AND tagPK = $3`, f.sitePK, f.typePK, f.tagPK); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func (t *dataLatencyTag) all(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT siteID, tag, typeID from data.latency_tag
				JOIN mtr.tag USING (tagpk)
				JOIN data.site USING (sitepk)
				JOIN data.type USING (typepk)
				ORDER BY tag ASC`); err != nil {
		return weft.InternalServerError(err)
	}
	defer rows.Close()

	var ts mtrpb.DataLatencyTagResult

	for rows.Next() {
		var t mtrpb.DataLatencyTag

		if err = rows.Scan(&t.SiteID, &t.Tag, &t.TypeID); err != nil {
			return weft.InternalServerError(err)
		}

		ts.Result = append(ts.Result, &t)
	}

	var by []byte
	if by, err = proto.Marshal(&ts); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	h.Set("Content-Type", "application/x-protobuf")

	return &weft.StatusOK
}
