package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type fieldType struct {
	typePK int
	Scale  float64 // used to scale the stored metric for display
	Name   string
	Unit   string // display unit after the metric has been multiplied by scale.
}

var fieldTypes = map[string]fieldType{
	"voltage": fieldType{
		typePK: 1,
		Scale:  0.001,
		Name:   "Voltage",
		Unit:   "V",
	},
	"clock": fieldType{
		typePK: 2,
		Scale:  1.0,
		Name:   "Clock Quality",
		Unit:   "%",
	},
	"satellites": fieldType{
		typePK: 3,
		Scale:  1.0,
		Name:   "Satellites Tracked",
		Unit:   "n",
	},
	"conn": fieldType{
		typePK: 4,
		Scale:  0.001,
		Name:   "Connectivity",
		Unit:   "ms",
	},
	"ping": fieldType{
		typePK: 5,
		Scale:  0.001,
		Name:   "ping",
		Unit:   "ms",
	},

	"disk.hd1": fieldType{
		typePK: 6,
		Scale:  1.0,
		Name:   "disk hd1",
		Unit:   "%",
	},
	"disk.hd2": fieldType{
		typePK: 7,
		Scale:  1.0,
		Name:   "disk hd2",
		Unit:   "%",
	},
	"disk.hd3": fieldType{
		typePK: 8,
		Scale:  1.0,
		Name:   "disk hd3",
		Unit:   "%",
	},
	"disk.hd4": fieldType{
		typePK: 9,
		Scale:  1.0,
		Name:   "disk hd4",
		Unit:   "%",
	},

	"centre": fieldType{
		typePK: 10,
		Scale:  1.0,
		Name:   "centre",
		Unit:   "mV",
	},

	"rf.signal": fieldType{
		typePK: 11,
		Scale:  1.0,
		Name:   "rf signal",
		Unit:   "dB",
	},
	"rf.noise": fieldType{
		typePK: 12,
		Scale:  1.0,
		Name:   "rf noise",
		Unit:   "dB",
	},

}

func loadFieldType(typeID string) (fieldType, *result) {

	if f, ok := fieldTypes[typeID]; ok {
		return f, &statusOK
	}

	return fieldType{}, badRequest("invalid type " + typeID)
}

func (f *fieldType) jsonV1(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	if by, err := json.Marshal(fieldTypes); err == nil {
		b.Write(by)
	} else {
		return internalServerError(err)
	}

	h.Set("Content-Type", "application/json;version=1")

	return &statusOK
}
