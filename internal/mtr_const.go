package internal

/*
 */
type ID int16

const (
	// HTTP requests
	Requests ID = 1

	// HTTP status codes (100 - 999).
	StatusOK                  ID = 200
	StatusBadRequest          ID = 400
	StatusUnauthorized        ID = 401
	StatusNotFound            ID = 404
	StatusInternalServerError ID = 500
	StatusServiceUnavailable  ID = 503

	// MemStats cf pkg/runtime/#MemStats
	// Also https://software.intel.com/en-us/blogs/2014/05/10/debugging-performance-issues-in-go-programs
	MemSys         ID = 1000 // bytes obtained from system
	MemHeapAlloc   ID = 1001 // bytes allocated and not yet freed
	MemHeapSys     ID = 1002 // bytes obtained from system
	MemHeapObjects ID = 1003 // total number of allocated objects

	// Other runtime stats
	Routines ID = 1100 // number of Go routines in use.

	// Messaging
	MsgRx   ID = 1201
	MsgTx   ID = 1202
	MsgProc ID = 1203
	MsgErr  ID = 1204
)

var idColours = map[int]string{
	1:   "lawngreen",
	200: "deepskyblue",
	400: "sandybrown",
	401: "tan",
	404: "lightcoral",
	500: "tomato",
	503: "tomato",

	1000: "lawngreen",
	1001: "deepskyblue",
	1002: "tan",
	1003: "deepskyblue",

	1100: "deepskyblue",

	1201: "deepskyblue",
	1202: "gold",
	1203: "deepskyblue",
	1204: "tomato",
}

var idLables = map[int]string{
	1:   "Requests",
	200: "200 OK",
	400: "400 Bad Request",
	401: "401 Unauthorized",
	404: "404 Not Found",
	500: "500 Internal Server Error",
	503: "503 Service Unavailable",

	1000: "Mem Sys",
	1001: "Mem Heap Alloc",
	1002: "Mem Heap Sys",
	1003: "Mem Heap Objects",

	1100: "Go Routines",

	1201: "Msg Rx",
	1202: "Msg Tx",
	1203: "Msg Processed",
	1204: "Msg Error",
}

func Colour(id int) string {
	if s, ok := idColours[id]; ok {
		return s
	}

	return "yellow"
}

func Lable(id int) string {
	if s, ok := idLables[id]; ok {
		return s
	}

	return "que"
}
