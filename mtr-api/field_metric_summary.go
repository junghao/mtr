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

// for SVG maps.
type point struct {
	latitude, longitude float64
	x, y                float64
}

// TODO: returns weft.NotFound when query result is empty?
func fieldLatestProto(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	typeID := r.URL.Query().Get("typeID")

	var err error
	var rows *sql.Rows

	switch typeID {
	case "":
		rows, err = dbR.Query(`select deviceID, modelID, typeid, time, value, lower, upper, scale
		FROM field.metric_summary
		JOIN field.device using (devicePK)
		JOIN field.model using (modelPK)
		JOIN field.threshold using (devicePK, typePK)
		JOIN field.type using (typePK)`)
	default:
		rows, err = dbR.Query(`select deviceID, modelID, typeid, time, value, lower, upper, scale
		FROM field.metric_summary
		JOIN field.device using (devicePK)
		JOIN field.model using (modelPK)
		JOIN field.threshold using (devicePK, typePK)
		JOIN field.type using (typePK)
		WHERE typeID = $1;`, typeID)
	}
	if err != nil {
		return weft.InternalServerError(err)
	}

	defer rows.Close()

	var t time.Time
	var fmlr mtrpb.FieldMetricSummaryResult

	for rows.Next() {
		var fmr mtrpb.FieldMetricSummary

		if err = rows.Scan(&fmr.DeviceID, &fmr.ModelID, &fmr.TypeID, &t, &fmr.Value,
			&fmr.Lower, &fmr.Upper, &fmr.Scale); err != nil {
			return weft.InternalServerError(err)
		}

		fmr.Seconds = t.Unix()

		fmlr.Result = append(fmlr.Result, &fmr)
	}
	rows.Close()

	var by []byte

	if by, err = proto.Marshal(&fmlr); err != nil {
		return weft.InternalServerError(err)
	}

	b.Write(by)

	return &weft.StatusOK
}

func fieldLatestSvg(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
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

	// TODO: handle maps that cross 180 (ST_Within)
	if rows, err = dbR.Query(`WITH p as (SELECT geom, time, value, lower, upper,
			ST_Transform(geom::geometry, 3857) as pt
			FROM field.metric_summary
			JOIN field.device using (devicePK)
			JOIN field.threshold using (devicePK, typePK)
			JOIN field.type using (typePK)
			WHERE typeID = $1)
			SELECT ST_X(pt), ST_Y(pt)*-1, ST_X(geom::geometry), ST_Y(geom::geometry), time, value, lower, upper FROM p
			WHERE ST_Within(geom::geometry, ST_GeomFromText($2, 4326))`, typeID, bboxWkt); err != nil {
		return weft.InternalServerError(err)
	}
	defer rows.Close()

	ago := time.Now().UTC().Add(time.Hour * -3)

	var late []point
	var good []point
	var bad []point
	var dunno []point

	for rows.Next() {
		var p point
		var t time.Time
		var min, max, v int

		if err = rows.Scan(&p.x, &p.y, &p.longitude, &p.latitude, &t, &v, &min, &max); err != nil {
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
		switch {
		case t.Before(ago):
			late = append(late, p)
		case min == 0 && max == 0:
			dunno = append(dunno, p)
		case v < min || v > max:
			bad = append(bad, p)
		default:
			good = append(good, p)
		}
	}
	rows.Close()

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

func fieldLatestGeoJSON(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	var rows *sql.Rows
	var err error

	typeID := r.URL.Query().Get("typeID")

	var d string
	err = db.QueryRow("select typeID FROM field.type where typeID = $1", typeID).Scan(&d)

	if err == sql.ErrNoRows {
		return &weft.NotFound
	}
	if err != nil {
		return weft.ServiceUnavailableError(err)
	}

	if rows, err = dbR.Query(`
		WITH p as (SELECT geom, time, value, lower, upper, deviceid, typeid
		FROM field.metric_summary
		JOIN field.device using (devicePK)
		JOIN field.threshold using (devicePK, typePK)
		JOIN field.type using (typePK)
		WHERE typeID = $1)
		SELECT row_to_json(fc)
		FROM ( SELECT 'FeatureCollection' as type, COALESCE(array_to_json(array_agg(f)), '[]') as features
		from (SELECT 'Feature' as type,
				ST_AsGeoJSON(p.geom)::json as geometry,
				row_to_json(
					(SELECT l FROM
						(
						SELECT
						"time",
						value,
						lower,
						upper,
						deviceid,
						typeid
						) as l
					)
				) as properties FROM p
		) as f ) as fc`, typeID); err != nil {
		return weft.InternalServerError(err)
	}
	defer rows.Close()

	var gj string

	if rows.Next() {

		if err = rows.Scan(&gj); err != nil {
			return weft.InternalServerError(err)
		}
	}

	rows.Close()
	b.WriteString(gj)

	return &weft.StatusOK
}
