package main

import (
	"github.com/lib/pq"
)

// http://www.postgresql.org/docs/9.4/static/errcodes-appendix.html
const (
	errorUniqueViolation pq.ErrorCode = "23505"
	DYGRAPH_TIME_FORMAT               = "2006/01/02 15:04:05"
)
