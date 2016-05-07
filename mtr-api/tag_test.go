package main

import (
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/golang/protobuf/proto"
	"testing"
)

func TestTag(t *testing.T) {
	setup(t)
	defer teardown()

	doRequest("PUT", "*/*", "/tag/TAUP", 200, t)
	doRequest("DELETE", "*/*", "/tag/TAUP", 200, t)
	doRequest("PUT", "*/*", "/tag/TAUP", 200, t)
}

func TestTagAllProto(t *testing.T) {
	setup(t)
	defer teardown()

	addFieldMetrics(t)
	addDataMetrics(t)

	// searches for all tags that have been added to a metric.
	// returns a protobus of tag strings for use in the type ahead etc.
	doRequest("GET", "application/x-protobuf", "/tag", 200, t)

	var b []byte
	var err error

	if b, err = getBytes("application/x-protobuf", "/tag"); err != nil {
		t.Error(err)
	}

	var tr mtrpb.TagResult

	if err = proto.Unmarshal(b, &tr); err != nil {
		t.Error(err)
	}

	if len(tr.Used) != 3 {
		t.Errorf("expected 3 active tags got %d", len(tr.Used))
	}

	if tr.Used[0].Tag != "DAGG" {
		t.Errorf("expected DAGG as the first tag got %s", tr.Used[0].Tag)
	}
}

func TestTagProto(t *testing.T) {
	setup(t)
	defer teardown()

	addFieldMetrics(t)
	addDataMetrics(t)

	// searches for all metrics with the tag TAUP and returns a protobuf
	// of summary results.
	doRequest("GET", "application/x-protobuf", "/tag/TAUP", 200, t)

	var b []byte
	var err error

	if b, err = getBytes("application/x-protobuf", "/tag/TAUP"); err != nil {
		t.Error(err)
	}

	var tr mtrpb.TagSearchResult

	if err = proto.Unmarshal(b, &tr); err != nil {
		t.Error(err)
	}

	if tr.FieldMetric == nil {
		t.Error("Got nil FieldMetric")
	}

	if tr.DataLatency == nil {
		t.Error("Got nil DataLatency")
	}

	if tr.FieldMetric[0].DeviceID != "gps-taupoairport" {
		t.Errorf("expected deviceID gps-taupoairport got %s", tr.FieldMetric[0].DeviceID)
	}

	if tr.DataLatency[0].SiteID != "TAUP" {
		t.Errorf("expected siteID TAUP got %s", tr.DataLatency[0].SiteID)
	}
}
