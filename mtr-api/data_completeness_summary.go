package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/map180"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/GeoNet/weft"
	"github.com/golang/protobuf/proto"
	"math"
	"net/http"
	"strconv"
	"time"
)

func dataCompletenessSummaryProto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	typeID := r.URL.Query().Get("typeID")

	var err error
	var rows *sql.Rows
	var expected int

	switch typeID {
	case "":
		rows, err = dbR.Query(`SELECT siteID, typeID, count, expected
		FROM data.completeness_summary
		JOIN data.site USING (sitePK)
		JOIN data.completeness_type USING (typePK)`)
	default:
		var typePK int
		if err = dbR.QueryRow(`SELECT typePK FROM data.completeness_type WHERE typeID = $1`,
			typeID).Scan(&typePK); err != nil {
			if err == sql.ErrNoRows {
				return &weft.NotFound
			}
			return weft.InternalServerError(err)
		}

		rows, err = dbR.Query(`SELECT siteID, typeID, count, expected
		FROM data.completeness_summary
		JOIN data.site USING (sitePK)
		JOIN data.completeness_type USING (typePK)
		WHERE typeID = $1;`, typeID)
	}

	if err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var dcr mtrpb.DataCompletenessSummaryResult

	for rows.Next() {
		var count int
		var siteID string

		if err = rows.Scan(&siteID, &typeID, &count, &expected); err != nil {
			return weft.InternalServerError(err)
		}

		c := float32(count) / (float32(expected) / 288)
		dc := mtrpb.DataCompletenessSummary{TypeID: typeID, SiteID: siteID, Completeness: c, Seconds: t.Unix()}
		dcr.Result = append(dcr.Result, &dc)
	}

	var by []byte

	if by, err = proto.Marshal(&dcr); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	return &weft.StatusOK
}

func dataCompletenessSummarySvg(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var rows *sql.Rows
	var width int
	var err error

	typeID := r.URL.Query().Get("typeID")
	bbox := r.URL.Query().Get("bbox")

	if err = map180.ValidBbox(bbox); err != nil {
		return weft.BadRequest(err.Error())
	}

	if width, err = strconv.Atoi(r.URL.Query().Get("width")); err != nil {
		return weft.BadRequest("invalid width")
	}

	var raw map180.Raw
	if raw, err = wm.MapRaw(bbox, width); err != nil {
		return weft.InternalServerError(err)
	}

	var bboxWkt string
	if bboxWkt, err = map180.BboxToWKTPolygon(bbox); err != nil {
		return weft.InternalServerError(err)
	}

	if rows, err = dbR.Query(`with p as (select geom, time, count, expected,
			st_transform(geom::geometry, 3857) as pt
			FROM data.completeness_summary
			JOIN data.site USING (sitePK)
			JOIN data.completeness_type USING (typePK)
			where typeID = $1)
			select ST_X(pt), ST_Y(pt)*-1, ST_X(geom::geometry),ST_Y(geom::geometry), time,
			count, expected from p
			WHERE ST_Within(geom::geometry, ST_GeomFromText($2, 4326))`, typeID, bboxWkt); err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	//ago := time.Now().UTC().Add(time.Hour * -3)

	var late []point
	var good []point
	var bad []point
	var dunno []point

	for rows.Next() {
		var p point
		var t time.Time
		var count int
		var expected int

		if err = rows.Scan(&p.x, &p.y, &p.longitude, &p.latitude, &t, &count, &expected); err != nil {
			return weft.InternalServerError(err)
		}

		// Does not handle crossing the equator.
		switch {
		case raw.CrossesCentral && p.longitude > -180.0 && p.longitude < 0.0:
			p.x = (p.x + map180.Width3857 - raw.LLX) * raw.DX
			p.y = (p.y - math.Abs(raw.YShift)) * raw.DX
		case p.longitude > 0.0:
			p.x = (p.x - math.Abs(raw.XShift)) * raw.DX
			p.y = (p.y - math.Abs(raw.YShift)) * raw.DX
		default:
			p.x = (p.x + math.Abs(raw.XShift)) * raw.DX
			p.y = (p.y - math.Abs(raw.YShift)) * raw.DX

		}

		// TODO: Define what is "Bad"
		completeness := float64(count) / (float64(expected) / 288)
		if completeness >= 1.0 {
			good = append(good, p)
		} else {
			bad = append(bad, p)
		}
	}

	b.WriteString(`<?xml version="1.0"?>`)
	b.WriteString(fmt.Sprintf("<svg  viewBox=\"0 0 %d %d\"  xmlns=\"http://www.w3.org/2000/svg\">",
		raw.Width, raw.Height))
	b.WriteString(fmt.Sprintf("<rect x=\"0\" y=\"0\" width=\"%d\" height=\"%d\" style=\"fill: azure\"/>", raw.Width, raw.Height))
	b.WriteString(fmt.Sprintf("<path style=\"fill: wheat; stroke-width: 1; stroke-linejoin: round; stroke: lightslategrey\" d=\"%s\"/>", raw.Land))
	b.WriteString(fmt.Sprintf("<path style=\"fill: azure; stroke-width: 1; stroke-linejoin: round; stroke: lightslategrey\" d=\"%s\"/>", raw.Lakes))

	b.WriteString("<g style=\"stroke: #377eb8; fill: #377eb8; \">") // blueish
	for _, p := range dunno {
		b.WriteString(fmt.Sprintf("<circle cx=\"%.1f\" cy=\"%.1f\" r=\"%d\"/>", p.x, p.y, 5))
	}
	b.WriteString("</g>")

	b.WriteString("<g style=\"stroke: #4daf4a; fill: #4daf4a; \">") // greenish
	for _, p := range good {
		b.WriteString(fmt.Sprintf("<circle cx=\"%.1f\" cy=\"%.1f\" r=\"%d\"/>", p.x, p.y, 5))
	}
	b.WriteString("</g>")

	b.WriteString("<g style=\"stroke: #e41a1c; fill: #e41a1c; \">") //red
	for _, p := range bad {
		b.WriteString(fmt.Sprintf("<circle cx=\"%.1f\" cy=\"%.1f\" r=\"%d\"/>", p.x, p.y, 6))
	}
	b.WriteString("</g>")

	b.WriteString("<g style=\"stroke: #984ea3; fill: #984ea3; \">") // purple
	for _, p := range late {
		b.WriteString(fmt.Sprintf("<circle cx=\"%.1f\" cy=\"%.1f\" r=\"%d\"/>", p.x, p.y, 6))
	}
	b.WriteString("</g>")

	b.WriteString("</svg>")

	return &weft.StatusOK
}
