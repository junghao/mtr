package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"github.com/lib/pq"
	"net/http"
)

func dataCompletenessTagPut(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	v := r.URL.Query()

	var err error
	var result sql.Result

	if result, err = db.Exec(`INSERT INTO data.completeness_tag(sitePK, typePK, tagPK)
				SELECT sitePK, typePK, tagPK
				FROM data.site, data.completeness_type, mtr.tag
				WHERE siteID = $1
				AND typeID = $2
				AND tag = $3`,
		v.Get("siteID"), v.Get("typeID"), v.Get("tag")); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
			return &weft.StatusOK
		} else {
			return weft.InternalServerError(err)
		}
	}

	var i int64
	if i, err = result.RowsAffected(); err != nil {
		return weft.InternalServerError(err)
	}
	if i != 1 {
		return weft.BadRequest("Didn't create row, check your query parameters exist")
	}

	return &weft.StatusOK
}

func dataCompletenessTagDelete(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	v := r.URL.Query()

	if _, err := db.Exec(`DELETE FROM data.completeness_tag
			WHERE sitePK = (SELECT sitePK FROM data.site WHERE siteID = $1)
			AND typePK = (SELECT typePK FROM data.completeness_type WHERE typeID = $2)
			AND tagPK = (SELECT tagPK FROM mtr.tag where tag = $3)`,
		v.Get("siteID"), v.Get("typeID"), v.Get("tag")); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func dataCompletenessTagProto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT siteID, tag, typeID from data.completeness_tag
				JOIN mtr.tag USING (tagpk)
				JOIN data.site USING (sitepk)
				JOIN data.completeness_type USING (typepk)
				ORDER BY tag ASC`); err != nil {
		return weft.InternalServerError(err)
	}
	defer rows.Close()

	var ts mtrpb.DataCompletenessTagResult

	for rows.Next() {
		var t mtrpb.DataCompletenessTag

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

	return &weft.StatusOK
}
