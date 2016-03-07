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
