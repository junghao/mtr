package main

import (
	"testing"
)

func TestAppMetricPK(t *testing.T) {
	setup(t)
	defer teardown()
	var i, j int
	var res *result

	if i, res = applicationPK("test-app"); !res.ok {
		t.Error(res.msg)
	}
	if i == 0 {
		t.Error("got 0 for applicationPK")
	}

	if j, res = applicationPK("test-app"); !res.ok {
		t.Error(res.msg)
	}
	if j == 0 {
		t.Error("got 0 for applicationPK")
	}

	if i != j {
		t.Error("applicationPK should be the same between calls.")
	}

	if i, res = instancePK("test-app"); !res.ok {
		t.Error(res.msg)
	}
	if i == 0 {
		t.Error("got 0 for instancePK")
	}

	if j, res = instancePK("test-app"); !res.ok {
		t.Error(res.msg)
	}
	if j == 0 {
		t.Error("got 0 for instancePK")
	}

	if i != j {
		t.Error("instancePK should be the same between calls.")
	}

	if i, res = sourcePK("test-app"); !res.ok {
		t.Error(res.msg)
	}
	if i == 0 {
		t.Error("got 0 for sourcePK")
	}

	if j, res = sourcePK("test-app"); !res.ok {
		t.Error(res.msg)
	}
	if j == 0 {
		t.Error("got 0 for sourcePK")
	}

	if i != j {
		t.Error("sourcePK should be the same between calls.")
	}

}
