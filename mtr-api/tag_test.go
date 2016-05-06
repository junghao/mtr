package main

import "testing"

func TestTag(t *testing.T) {
	setup(t)
	defer teardown()

	doRequest("PUT", "*/*", "/tag?tag=TAUP", 200, t)
	doRequest("DELETE", "*/*", "/tag?tag=TAUP", 200, t)
	doRequest("PUT", "*/*", "/tag?tag=TAUP", 200, t)
}
