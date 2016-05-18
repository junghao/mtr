package main

import (
	"strings"
)

var colours = [9]string{
	"#a6cee3",
	"#1f78b4",
	"#b2df8a",
	"#33a02c",
	"#fb9a99",
	"#e31a1c",
	"#fdbf6f",
	"#ff7f00",
	"#cab2d6",
}

var blues = [9]string{
	"#023858",
	"#045a8d",
	"#0570b0",
	"#3690c0",
	"#74a9cf",
	"#a6bddb",
	"#d0d1e6",
	"#ece7f2",
	"#fff7fb",
}

var browns = [9]string{
	"#662506",
	"#993404",
	"#cc4c02",
	"#ec7014",
	"#fe9929",
	"#fec44f",
	"#fee391",
	"#fff7bc",
	"#ffffe5",
}

var purples = [9]string{
	"#fff7f3",
	"#fde0dd",
	"#fcc5c0",
	"#fa9fb5",
	"#f768a1",
	"#dd3497",
	"#ae017e",
	"#7a0177",
	"#49006a",
}

func svgColour(id string, pk int) string {
	if pk < 0 {
		return "yellow"
	}

	// pk is typically small.  May need a different way of choosing color later.
	for pk > 9 {
		pk = pk - 10
	}

	i := strings.LastIndex(id, ".")
	if i > -1 && i+1 < len(id) {
		id = id[i+1:]
	}

	switch id {
	case "GET":
		return blues[pk]
	case "PUT":
		return browns[pk]
	case "DELETE":
		return purples[pk]
	default:
		return colours[pk]
	}
}
