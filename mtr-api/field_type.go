package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/GeoNet/weft"
	"net/http"
)

var fieldTypes = map[string]fieldType{
	"voltage": {
		pk:    1,
		Scale: 0.001,
		Name:  "Voltage",
		Unit:  "V",
	},
	"clock": {
		pk:    2,
		Scale: 1.0,
		Name:  "Clock Quality",
		Unit:  "%",
	},
	"satellites": {
		pk:    3,
		Scale: 1.0,
		Name:  "Satellites Tracked",
		Unit:  "n",
	},
	"conn": {
		pk:    4,
		Scale: 0.001,
		Name:  "Connectivity",
		Unit:  "ms",
	},
	"ping": {
		pk:    5,
		Scale: 0.001,
		Name:  "ping",
		Unit:  "ms",
	},

	"disk.hd1": {
		pk:    6,
		Scale: 1.0,
		Name:  "disk hd1",
		Unit:  "%",
	},
	"disk.hd2": {
		pk:    7,
		Scale: 1.0,
		Name:  "disk hd2",
		Unit:  "%",
	},
	"disk.hd3": {
		pk:    8,
		Scale: 1.0,
		Name:  "disk hd3",
		Unit:  "%",
	},
	"disk.hd4": {
		pk:    9,
		Scale: 1.0,
		Name:  "disk hd4",
		Unit:  "%",
	},

	"centre": {
		pk:    10,
		Scale: 1.0,
		Name:  "centre",
		Unit:  "mV",
	},

	"rf.signal": {
		pk:    11,
		Scale: 1.0,
		Name:  "rf signal",
		Unit:  "dB",
	},
	"rf.noise": {
		pk:    12,
		Scale: 1.0,
		Name:  "rf noise",
		Unit:  "dB",
	},
}

func loadFieldType(typeID string) (fieldType, *weft.Result) {

	if f, ok := fieldTypes[typeID]; ok {
		return f, &weft.StatusOK
	}

	return fieldType{}, weft.BadRequest("invalid type " + typeID)
}

func (a *fieldType) read(typeID string) *weft.Result {
	if typeID == "" {
		return weft.InternalServerError(fmt.Errorf("empty typeID"))
	}

	a.id = typeID

	var t fieldType
	var res *weft.Result

	if t, res = loadFieldType(a.id); !res.Ok {
		return res
	}

	a.pk = t.pk
	a.Scale = t.Scale
	a.Name = t.Name
	a.Unit = t.Unit

	return &weft.StatusOK
}

func (f *fieldType) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	if by, err := json.Marshal(fieldTypes); err == nil {
		b.Write(by)
	} else {
		return weft.InternalServerError(err)
	}

	h.Set("Content-Type", "application/json;version=1")

	return &weft.StatusOK
}
