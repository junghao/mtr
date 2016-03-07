package main

import (
	"github.com/lib/pq"
)

// http://www.postgresql.org/docs/9.4/static/errcodes-appendix.html
const errorUniqueViolation pq.ErrorCode = "23505"
