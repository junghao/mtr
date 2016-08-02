package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"net/http"
)

func dataTypeProto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT typeID, display FROM data.type ORDER BY typeID ASC`); err != nil {
		return weft.InternalServerError(err)
	}
	defer rows.Close()

	var ftr mtrpb.DataTypeResult

	for rows.Next() {
		var ft mtrpb.DataType

		if err = rows.Scan(&ft.TypeID, &ft.Display); err != nil {
			return weft.InternalServerError(err)
		}

		ftr.Result = append(ftr.Result, &ft)
	}

	var by []byte
	if by, err = proto.Marshal(&ftr); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	return &weft.StatusOK
}

func dataCompletenessTypeProto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT typeID FROM data.completeness_type ORDER BY typeID ASC`); err != nil {
		return weft.InternalServerError(err)
	}
	defer rows.Close()

	var ftr mtrpb.DataTypeResult

	for rows.Next() {
		var ft mtrpb.DataType

		if err = rows.Scan(&ft.TypeID); err != nil {
			return weft.InternalServerError(err)
		}

		ftr.Result = append(ftr.Result, &ft)
	}

	var by []byte
	if by, err = proto.Marshal(&ftr); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	return &weft.StatusOK
}
