package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"net/http"
)

type appID struct {
}

// write a protobuf to b of all applicationid's in app.application
func (a appID) proto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	var err error
	var rows *sql.Rows

	if rows, err = dbR.Query(`SELECT applicationid FROM app.application ORDER BY applicationid ASC`); err != nil {
		return weft.InternalServerError(err)
	}
	defer rows.Close()

	var ar mtrpb.AppIDSummaryResult

	for rows.Next() {
		var ai mtrpb.AppIDSummary

		if err = rows.Scan(&ai.ApplicationID); err != nil {
			return weft.InternalServerError(err)
		}

		ar.Result = append(ar.Result, &ai)
	}

	var by []byte

	if by, err = proto.Marshal(&ar); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	h.Set("Content-Type", "application/x-protobuf")

	return &weft.StatusOK
}
