package main

import (
	"github.com/GeoNet/weft"
)

var dataTypes = map[string]dataType{
	"latency.strong": {
		pk:    1,
		Scale: 1.0,
		Name:  "latency strong motion data",
		Unit:  "ms",
	},
	"latency.weak": {
		pk:    2,
		Scale: 1.0,
		Name:  "latency weak motion data",
		Unit:  "ms",
	},
	"latency.gnss.1hz": {
		pk:    3,
		Scale: 1.0,
		Name:  "latency GNSS 1Hz data",
		Unit:  "ms",
	},
	"latency.tsunami": {
		pk:    4,
		Scale: 1.0,
		Name:  "latency tsunami data",
		Unit:  "ms",
	},
}

func (d *dataType) read(typeID string) *weft.Result {
	if typeID == "" {
		return weft.BadRequest("empty typeID")
	}

	d.id = typeID

	var t dataType
	var ok bool
	if t, ok = dataTypes[d.id]; !ok {
		return weft.BadRequest("invalid typeID " + d.id)
	}

	d.pk = t.pk
	d.Scale = t.Scale
	d.Name = t.Name
	d.Unit = t.Unit
	return &weft.StatusOK
}

