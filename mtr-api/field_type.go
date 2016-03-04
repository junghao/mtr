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
