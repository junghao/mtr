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

func fieldModelPut(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if _, err := db.Exec(`INSERT INTO field.model(modelID) VALUES($1)`, r.URL.Query().Get("modelID")); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errorUniqueViolation {
			// ignore unique constraint errors
		} else {
			return weft.InternalServerError(err)
		}
	}

	return &weft.StatusOK
}

func fieldModelDelete(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if _, err := db.Exec(`DELETE FROM field.model where modelID = $1`, r.URL.Query().Get("modelID")); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func fieldModelProto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT modelID
		FROM
		field.model`); err != nil {
		return weft.InternalServerError(err)
	}

	var fmr mtrpb.FieldModelResult

	for rows.Next() {
		var t mtrpb.FieldModel

		if err = rows.Scan(&t.ModelID); err != nil {
			return weft.InternalServerError(err)
		}

		fmr.Result = append(fmr.Result, &t)
	}

	var by []byte
	if by, err = proto.Marshal(&fmr); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	return &weft.StatusOK
}
