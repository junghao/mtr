package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"net/http"
	"time"
)

type dataLatencySummary struct {
}

func (d dataLatencySummary) get(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{"typeID"}); !res.Ok {
		return res
	}

	typeID := r.URL.Query().Get("typeID")

	var err error
	var rows *sql.Rows

	switch typeID {
	case "":
		rows, err = dbR.Query(`SELECT siteID, typeID, time, mean, fifty, ninety, lower, upper
		FROM data.latency_summary`)
	default:
		rows, err = dbR.Query(`SELECT siteID, typeID, time, mean, fifty, ninety, lower, upper
		FROM data.latency_summary
		WHERE typeID = $1;`, typeID)
	}
	if err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var dlsr mtrpb.DataLatencySummaryResult

	for rows.Next() {

		var dls mtrpb.DataLatencySummary

		if err = rows.Scan(&dls.SiteID, &dls.TypeID, &t, &dls.Mean, &dls.Fifty, &dls.Ninety,
			&dls.Lower, &dls.Upper); err != nil {
			return weft.InternalServerError(err)
		}

		dls.Seconds = t.Unix()

		dlsr.Result = append(dlsr.Result, &dls)
	}
	rows.Close()

	var by []byte

	if by, err = proto.Marshal(&dlsr); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	h.Set("Content-Type", "application/x-protobuf")

	return &weft.StatusOK
}
