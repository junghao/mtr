package main

import (
	"fmt"
	"github.com/GeoNet/mtr/mtrpb"
	wt "github.com/GeoNet/weft/wefttest"
	"github.com/golang/protobuf/proto"
	"net/http"
	"testing"
	"time"
)

func init() {
	for i, _ := range routes {
		switch routes[i].Method {
		case "", "GET":
			// default Surrogate-Control cache times.
			if routes[i].Surrogate == "" {
				routes[i].Surrogate = "max-age=10"
			}
		default:
			// Any non GET requests need authentication
			routes[i].User = userW
			routes[i].Password = keyW
		}
	}
}

// routes test the API.
var routes = wt.Requests{
	// application metrics

	// all metrics for test-app are deleted in setup()

	// add a metric value
	{ID: wt.L(), URL: "/application/metric?applicationID=test-app&instanceID=test-instance&typeID=1000&value=10000&time=2015-05-14T21:40:30Z", Method: "PUT"},

	// add a counter value
	{ID: wt.L(), URL: "/application/counter?applicationID=test-app&typeID=200&count=10&time=2015-05-14T21:40:30Z", Method: "PUT"},

	// Another counter value at the same time increments the count.
	{ID: wt.L(), URL: "/application/counter?applicationID=test-app&typeID=200&count=10&time=2015-05-14T21:40:30Z", Method: "PUT"},

	// Add a timer value.  TODO - repeat at the same time errors.
	{ID: wt.L(), URL: "/application/timer?applicationID=test-app&sourceID=func-name&count=10&average=12&fifty=13&ninety=14&time=2015-05-14T21:40:30Z", Method: "PUT"},
	{ID: wt.L(), URL: "/application/timer?applicationID=test-app&sourceID=func-name&count=10&average=12&fifty=13&ninety=14&time=2015-05-14T21:40:30Z", Method: "PUT", Status: 500},

	// SVG plots
	{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=timers"},
	{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=counters"},
	{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=memory"},
	{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=objects"},
	{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=routines"},

	// field metrics

	// Creates a device model.  Repeated requests noop.
	{ID: wt.L(), URL: "/field/model?modelID=Trimble+NetR9", Method: "PUT"},

	// Delete a model then recreate it.  Delete cascades to devices
	// with that model (which cascades to metrics)
	{ID: wt.L(), URL: "/field/model?modelID=Trimble+NetR9", Method: "DELETE"},
	{ID: wt.L(), URL: "/field/model?modelID=Trimble+NetR9", Method: "PUT"},

	// Devices are at a lat long
	{ID: wt.L(), URL: "/field/device?deviceID=gps-taupoairport&modelID=Trimble+NetR9&latitude=-38.74270&longitude=176.08100", Method: "PUT"},
	{ID: wt.L(), URL: "/field/device?deviceID=gps-taupoairport", Method: "DELETE"},
	{ID: wt.L(), URL: "/field/device?deviceID=gps-taupoairport&modelID=Trimble+NetR9&latitude=-38.74270&longitude=176.08100", Method: "PUT"},

	// Delete all metrics typeID for a device
	{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage", Method: "DELETE"},

	// Save a metrics
	{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage&time=2015-05-14T21:40:30Z&value=14100", Method: "PUT"},

	// Should get a rate limit error for sends in the same minute
	{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage&time=2015-05-14T21:40:30Z&value=15100", Method: "PUT", Status: http.StatusTooManyRequests},

	// Tags
	{ID: wt.L(), URL: "/tag/LINZ", Method: "DELETE"},

	// tag must exist before it can be added to a metric
	{ID: wt.L(), URL: "/tag/LINZ", Method: "PUT"},
	{ID: wt.L(), URL: "/tag/TAUP", Method: "PUT"},

	// Create a tag on a metric type.  Multiple tags per metric are possible.  Repeat PUT is ok.
	{ID: wt.L(), URL: "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=TAUP", Method: "PUT"},
	{ID: wt.L(), URL: "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", Method: "PUT"},

	// Delete a tag on a metric
	{ID: wt.L(), URL: "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", Method: "DELETE"},

	// Tags
	{ID: wt.L(), URL: "/tag/LINZ", Method: "DELETE"},

	// tag must exist before it can be added to a metric
	{ID: wt.L(), URL: "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", Method: "PUT", Status: http.StatusBadRequest},
	{ID: wt.L(), URL: "/tag/LINZ", Method: "PUT"},
	{ID: wt.L(), URL: "/tag/TAUP", Method: "PUT"},

	// Create a tag on a metric type.  Multiple tags per metric are possible.  Repeat PUT is ok.
	{ID: wt.L(), URL: "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=TAUP", Method: "PUT"},
	{ID: wt.L(), URL: "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", Method: "PUT"},

	// Delete a tag on a metric
	{ID: wt.L(), URL: "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", Method: "DELETE"},

	// protobuf of all tagged field metrics
	{ID: wt.L(), URL: "/field/metric/tag", Accept: "application/x-protobuf"},

	// Thresholds
	// Create a threshold on a metric
	{ID: wt.L(), URL: "/field/metric/threshold?deviceID=gps-taupoairport&typeID=voltage&lower=12000&upper=15000", Method: "PUT"},

	// Update a threshold on a metric
	{ID: wt.L(), URL: "/field/metric/threshold?deviceID=gps-taupoairport&typeID=voltage&lower=13000&upper=15000", Method: "PUT"},

	// Delete a threshold on a metric then create it again
	{ID: wt.L(), URL: "/field/metric/threshold?deviceID=gps-taupoairport&typeID=voltage", Method: "DELETE"},
	{ID: wt.L(), URL: "/field/metric/threshold?deviceID=gps-taupoairport&typeID=voltage&lower=12000&upper=45000", Method: "PUT"},

	// GET requests
	// Non specific Accept headers return svg.
	// Model
	{ID: wt.L(), URL: "/field/model", Accept: "application/json;version=1"},

	// Device
	{ID: wt.L(), URL: "/field/device", Accept: "application/json;version=1"},

	// Metrics.  Resolution is optional on plots.  Resolution is fixed for sparks.
	// Options for the plot parameter:
	// line [default] = line plot.
	// spark = spark line
	{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage", Content: "image/svg+xml"},
	{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=minute", Content: "image/svg+xml"},
	{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=five_minutes", Content: "image/svg+xml"},
	{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=hour", Content: "image/svg+xml"},
	{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=minute", Content: "image/svg+xml"},
	{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=hour", Content: "image/svg+xml"},
	{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=day", Status: http.StatusBadRequest, Surrogate: "max-age=86400"},
	{ID: wt.L(), URL: "/field/metric?deviceID=gps-taupoairport&typeID=voltage&plot=spark", Content: "image/svg+xml"},

	// Latest metrics as SVG map
	//  These only pass with the map180 data in the DB.
	// Values for bbox and insetBbox are ChathamIsland LakeTaupo NewZealand NewZealandRegion
	// RaoulIsland WhiteIsland
	// or lhe bounding box for the map defining the lower left and upper right longitude
	// latitude (EPSG:4327) corners e.g., <code>165,-48,179,-34</code>.  Latitude must be in the range -85 to 85.  Maps can be 180 centric and bbox
	// definitions for longitude can be -180 to 180 or 0 to 360
	//
	// {ID: wt.L(), URL: "/field/metric/summary?bbox=WhiteIsland&width=800&typeID=voltage", Content: "image/svg+xml"},
	// {ID: wt.L(), URL: "/field/metric/summary?bbox=NewZealand&width=800&typeID=voltage", Content: "image/svg+xml"}.

	// All latest metrics as a FieldMetricLatestResult protobuf
	{ID: wt.L(), URL: "/field/metric/summary", Accept: "application/x-protobuf"},

	// Latest voltage metrics
	{ID: wt.L(), URL: "/field/metric/summary?typeID=voltage", Accept: "application/x-protobuf"},

	// Thresholds
	{ID: wt.L(), URL: "/field/metric/threshold", Accept: "application/json;version=1"},

	// All field metric thresholds as protobuf
	{ID: wt.L(), URL: "/field/metric/threshold", Accept: "application/x-protobuf"},

	// Metric types
	{ID: wt.L(), URL: "/field/type", Accept: "application/json;version=1"},

	// Data latency

	// Delete site - cascades to latency values
	{ID: wt.L(), URL: "/data/site?siteID=TAUP", Method: "DELETE"},

	// create a site.  Lat lon are indicative only and may not be suitable for
	// precise data use.
	{ID: wt.L(), URL: "/data/site?siteID=TAUP&latitude=-38.74270&longitude=176.08100", Method: "PUT"},

	// update the site location
	{ID: wt.L(), URL: "/data/site?siteID=TAUP&latitude=-38.64270&longitude=176.08100", Method: "PUT"},

	// delete then recreate
	{ID: wt.L(), URL: "/data/site?siteID=TAUP", Method: "DELETE"},
	{ID: wt.L(), URL: "/data/site?siteID=TAUP&latitude=-38.74270&longitude=176.08100", Method: "PUT"},

	// Should get a rate limit error for sends in the same minute
	{ID: wt.L(), URL: "/data/latency?siteID=TAUP&typeID=latency.strong&time=2015-05-14T21:40:30Z&mean=10000", Method: "PUT"},
	{ID: wt.L(), URL: "/data/latency?siteID=TAUP&typeID=latency.strong&time=2015-05-14T21:40:30Z&mean=14100", Status: http.StatusTooManyRequests, Method: "PUT"},

	// Add another site, some latency data, then delete.
	{ID: wt.L(), URL: "/data/site?siteID=WGTN", Method: "DELETE"},
	{ID: wt.L(), URL: "/data/site?siteID=WGTN&latitude=-38.74270&longitude=176.08100", Method: "PUT"},

	// All data sites as protobuf
	{ID: wt.L(), URL: "/data/site", Accept: "application/x-protobuf"},

	// min, max, fifty, ninety are optional latency values
	{ID: wt.L(), URL: "/data/latency?siteID=WGTN&typeID=latency.strong&time=2015-05-14T23:40:30Z&mean=10000&min=10&max=100000&fifty=9000&ninety=12000", Method: "PUT"},

	{ID: wt.L(), URL: "/data/latency?siteID=WGTN&typeID=latency.strong", Method: "DELETE"},

	// Create a threshold for latency.
	// I assume a single threshold would be for mean, fifty, and ninety?
	{ID: wt.L(), URL: "/data/latency/threshold?siteID=TAUP&typeID=latency.strong", Method: "DELETE"},
	{ID: wt.L(), URL: "/data/latency/threshold?siteID=TAUP&typeID=latency.strong&lower=12000&upper=15000", Method: "PUT"},

	// Update a threshold
	{ID: wt.L(), URL: "/data/latency/threshold?siteID=TAUP&typeID=latency.strong&lower=13000&upper=15000", Method: "PUT"},

	// Delete a threshold then create it again
	{ID: wt.L(), URL: "/data/latency/threshold?siteID=TAUP&typeID=latency.strong", Method: "DELETE"},
	{ID: wt.L(), URL: "/data/latency/threshold?siteID=TAUP&typeID=latency.strong&lower=12000&upper=15000", Method: "PUT"},

	// protobuf of all latency thresholds
	{ID: wt.L(), URL: "/data/latency/threshold", Accept: "application/x-protobuf"},

	// Tags

	{ID: wt.L(), URL: "/tag/FRED", Method: "DELETE"},
	{ID: wt.L(), URL: "/tag/DAGG", Method: "DELETE"},

	// tag must exist before it can be added to a metric
	{ID: wt.L(), URL: "/data/latency/tag?siteID=FRED&typeID=latency.strong&tag=TAUP", Status: http.StatusBadRequest, Method: "PUT"},

	{ID: wt.L(), URL: "/tag/FRED", Method: "PUT"},
	{ID: wt.L(), URL: "/tag/DAGG", Method: "PUT"},
	{ID: wt.L(), URL: "/tag/TAUP", Method: "PUT"},

	// Create a tag on a latency.  Multiple tags per metric are possible.  Repeat PUT is ok.
	{ID: wt.L(), URL: "/data/latency/tag?siteID=TAUP&typeID=latency.strong&tag=FRED", Method: "PUT"},
	{ID: wt.L(), URL: "/data/latency/tag?siteID=TAUP&typeID=latency.strong&tag=DAGG", Method: "PUT"},
	{ID: wt.L(), URL: "/data/latency/tag?siteID=TAUP&typeID=latency.strong&tag=TAUP", Method: "PUT"},

	// protobuf of all tagged data latencies
	{ID: wt.L(), URL: "/data/latency/tag", Accept: "application/x-protobuf"},

	// Latency plots.  Resolution is optional on plots and sparks.
	// Options for the plot parameter:
	// line [default] = line plot.
	// spark = spark line.
	{ID: wt.L(), URL: "/data/latency?siteID=TAUP&typeID=latency.strong"},
	{ID: wt.L(), URL: "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=minute"},
	{ID: wt.L(), URL: "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=five_minutes"},
	{ID: wt.L(), URL: "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=hour"},
	{ID: wt.L(), URL: "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=minute"},
	{ID: wt.L(), URL: "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=hour"},
	{ID: wt.L(), URL: "/data/latency?siteID=TAUP&typeID=latency.strong&plot=spark"},

	// Tags

	{ID: wt.L(), URL: "/tag/LINZ", Method: "DELETE"},

	// tag must exist before it can be added to a metric
	{ID: wt.L(), URL: "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", Status: http.StatusBadRequest, Method: "PUT"},

	{ID: wt.L(), URL: "/tag/LINZ", Method: "PUT"},
	{ID: wt.L(), URL: "/tag/TAUP", Method: "PUT"},

	// Create a tag on a metric type.  Multiple tags per metric are possible.  Repeat PUT is ok.
	{ID: wt.L(), URL: "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=TAUP", Method: "PUT"},
	{ID: wt.L(), URL: "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", Method: "PUT"},

	// Delete a tag on a metric
	{ID: wt.L(), URL: "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", Method: "DELETE"},
}

// Test all routes give the expected response.  Also check with
// cache busters and extra query paramters.
func TestRoutes(t *testing.T) {
	setup(t)
	defer teardown()

	for _, r := range routes {
		if b, err := r.Do(testServer.URL); err != nil {
			t.Error(err)
			t.Error(string(b))
		}
	}

	if err := routes.DoCheckQuery(testServer.URL); err != nil {
		t.Error(err)
	}
}

// Any routes that are not GET should http.StatusUnauthorized without authorisation.
func TestRoutesNoAuth(t *testing.T) {
	setup(t)
	defer teardown()

	for _, r := range routes {
		switch r.Method {
		case "", "GET":
		default:
			r.User = ""
			r.Password = ""
			r.Status = http.StatusUnauthorized

			if _, err := r.Do(testServer.URL); err != nil {
				t.Error(err)
			}
		}
	}
}

/*
Adds some plot data. Run just this test:

    go test -run TestPlotData

Then visit

http://localhost:8080/field/metric?deviceID=gps-taupoairport&typeID=voltage
http://localhost:8080/data/latency?siteID=TAUP&typeID=latency.strong
*/
func TestPlotData(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	r := wt.Request{
		User:     userW,
		Password: keyW,
		Method:   "PUT",
	}
	var err error

	// Load some field metrics (every 5 mins)
	now := time.Now().UTC()
	v := 14000
	for i := -720; i < 0; i += 5 {
		if i >= -100 {
			v = int(14000*(1/(float64(i)+101.0))) + 10000
			if v > 14000 {
				v = 14000
			}
		}

		r.URL = fmt.Sprintf("/field/metric?deviceID=gps-taupoairport&typeID=voltage&time=%s&value=%d",
			now.Add(time.Duration(i)*time.Minute).Format(time.RFC3339), v)

		if _, err = r.Do(testServer.URL); err != nil {
			t.Error(err)
		}

	}

	// Load some latency metrics (every 5 mins)
	now = time.Now().UTC()
	v = 14000
	for i := -720; i < 0; i += 5 {
		if i >= -100 {
			v = int(14000*(1/(float64(i)+101.0))) + 10000
			if v > 14000 {
				v = 14000
			}
		}

		r.URL = fmt.Sprintf("/data/latency?siteID=TAUP&typeID=latency.strong&time=%s&mean=%d",
			now.Add(time.Duration(i)*time.Minute).Format(time.RFC3339), v)

		if _, err = r.Do(testServer.URL); err != nil {
			t.Error(err)
		}
	}
}

// All field metric tags as a protobuf.
func TestFieldMetricTag(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	r := wt.Request{ID: wt.L(), URL: "/field/metric/tag", Accept: "application/x-protobuf"}

	var b []byte
	var err error

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	var tr mtrpb.FieldMetricTagResult

	if err = proto.Unmarshal(b, &tr); err != nil {
		t.Error(err)
	}

	if tr.Result == nil {
		t.Error("got nil for /field/metric/tag protobuf")
	}

	if len(tr.Result) != 1 {
		t.Errorf("expected 1 tag result got %d", len(tr.Result))
	}

	if tr.Result[0].Tag != "TAUP" {
		t.Errorf("expected TAUP as the first tag got %s", tr.Result[0].Tag)
	}

	if tr.Result[0].DeviceID != "gps-taupoairport" {
		t.Errorf("expected gps-taupoairport as the first deviceID got %s", tr.Result[0].DeviceID)
	}

	if tr.Result[0].TypeID != "voltage" {
		t.Errorf("expected voltage as the first typeID got %s", tr.Result[0].TypeID)
	}
}

// protobuf of field metrics and latencies for a single tag.
func TestTag(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	if err := refreshViews(); err != nil {
		t.Error(err)
	}

	r := wt.Request{ID: wt.L(), URL: "/tag/TAUP", Accept: "application/x-protobuf"}

	var b []byte
	var err error

	if b, err = r.Do(testServer.URL); err != nil {
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

// all tags as a protobuf
func TestTagAll(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	r := wt.Request{ID: wt.L(), URL: "/tag", Accept: "application/x-protobuf"}

	var b []byte
	var err error

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	var tr mtrpb.TagResult

	if err = proto.Unmarshal(b, &tr); err != nil {
		t.Error(err)
	}

	if len(tr.Result) != 4 {
		t.Errorf("expected 4 tags got %d", len(tr.Result))
	}

	if tr.Result[0].Tag != "DAGG" {
		t.Errorf("expected DAGG as the first tag got %s", tr.Result[0].Tag)
	}
}

// protobuf of latency summary info.
func TestDataLatencySummary(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	if err := refreshViews(); err != nil {
		t.Error(err)
	}

	r := wt.Request{ID: wt.L(), URL: "/data/latency/summary", Accept: "application/x-protobuf"}

	var b []byte
	var err error

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	var f mtrpb.DataLatencySummaryResult

	if err = proto.Unmarshal(b, &f); err != nil {
		t.Error(err)
	}

	if len(f.Result) != 1 {
		t.Error("expected 1 result.")
	}

	d := f.Result[0]

	if d.SiteID != "TAUP" {
		t.Errorf("expected TAUP got %s", d.SiteID)
	}

	if d.TypeID != "latency.strong" {
		t.Errorf("expected latency.strong got %s", d.TypeID)
	}

	if d.Mean != 10000 {
		t.Errorf("expected 10000 got %d", d.Mean)
	}

	if d.Fifty != 0 {
		t.Errorf("expected 0 got %d", d.Fifty)
	}

	if d.Ninety != 0 {
		t.Errorf("expected 0 got %d", d.Ninety)
	}

	if d.Seconds == 0 {
		t.Error("unexpected zero seconds")
	}

	if d.Upper != 15000 {
		t.Errorf("expected 15000 got %d", d.Upper)
	}

	if d.Lower != 12000 {
		t.Errorf("expected 12000 got %d", d.Lower)
	}

	r.URL = "/data/latency/summary?typeID=latency.strong"

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	f.Reset()

	if err = proto.Unmarshal(b, &f); err != nil {
		t.Error(err)
	}

	if len(f.Result) != 1 {
		t.Error("expected 1 result.")
	}
}

// protobuf of field metric summary info.
func TestFieldMetricsSummary(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	if err := refreshViews(); err != nil {
		t.Error(err)
	}

	r := wt.Request{ID: wt.L(), URL: "/field/metric/summary", Accept: "application/x-protobuf"}

	var b []byte
	var err error

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	var f mtrpb.FieldMetricSummaryResult

	if err = proto.Unmarshal(b, &f); err != nil {
		t.Error(err)
	}

	if len(f.Result) != 1 {
		t.Error("expected 1 result.")
	}

	d := f.Result[0]

	if d.DeviceID != "gps-taupoairport" {
		t.Errorf("expected gps-taupoairport got %s", d.DeviceID)
	}

	if d.ModelID != "Trimble NetR9" {
		t.Errorf("expected Trimble NetR9 got %s", d.ModelID)
	}

	if d.TypeID != "voltage" {
		t.Errorf("expected voltage got %s", d.TypeID)
	}

	if d.Value != 14100 {
		t.Errorf("expected 14100 got %d", d.Value)
	}

	if d.Seconds == 0 {
		t.Error("unexpected zero seconds")
	}

	if d.Upper != 45000 {
		t.Errorf("expected 45000 got %d", d.Upper)
	}

	if d.Lower != 12000 {
		t.Errorf("expected 12000 got %d", d.Lower)
	}

	// should be no errors and empty result for typeID=conn
	r.URL = "/field/metric/summary?typeID=conn"

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	f.Reset()

	if err = proto.Unmarshal(b, &f); err != nil {
		t.Error(err)
	}

	if len(f.Result) != 0 {
		t.Error("expected 0 results.")
	}
}

// protobuf of field metric threshold info.
func TestFieldMetricsThreshold(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	r := wt.Request{ID: wt.L(), URL: "/field/metric/threshold", Accept: "application/x-protobuf"}

	var b []byte
	var err error

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	var f mtrpb.FieldMetricThresholdResult

	if err = proto.Unmarshal(b, &f); err != nil {
		t.Error(err)
	}

	if f.Result == nil {
		t.Error("got nil for /field/metric/threshold protobuf")
	}

	if len(f.Result) != 1 {
		t.Error("expected 1 result.")
	}

	d := f.Result[0]

	if d.DeviceID != "gps-taupoairport" {
		t.Errorf("expected gps-taupoairport got %s", d.DeviceID)
	}

	if d.TypeID != "voltage" {
		t.Errorf("expected voltage got %s", d.TypeID)
	}

	if d.Upper != 45000 {
		t.Errorf("expected 45000 got %d", d.Upper)
	}

	if d.Lower != 12000 {
		t.Errorf("expected 12000 got %d", d.Lower)
	}
}

// protobuf of data sites.
func TestDataSites(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	r := wt.Request{ID: wt.L(), URL: "/data/site", Accept: "application/x-protobuf"}

	var b []byte
	var err error

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	var f mtrpb.DataSiteResult

	if err = proto.Unmarshal(b, &f); err != nil {
		t.Error(err)
	}

	if f.Result == nil {
		t.Error("got nil for /data/site protobuf")
	}

	if len(f.Result) != 2 {
		t.Error("expected 2 results.")
	}

	var found bool

	for _, v := range f.Result {
		if v.SiteID == "TAUP" {
			found = true

			if v.Latitude != -38.74270 {
				t.Errorf("Data site TAUP got expected latitude -38.74270 got %f", v.Latitude)
			}

			if v.Longitude != 176.08100 {
				t.Errorf("Data site TAUP got expected longitude 176.08100 got %f", v.Longitude)
			}
		}
	}

	if !found {
		t.Error("Didn't find site TAUP")
	}
}

// All data latency tags as a protobuf.
func TestDataLatencyTag(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	r := wt.Request{ID: wt.L(), URL: "/data/latency/tag", Accept: "application/x-protobuf"}

	var b []byte
	var err error

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	var tr mtrpb.DataLatencyTagResult

	if err = proto.Unmarshal(b, &tr); err != nil {
		t.Error(err)
	}

	if tr.Result == nil {
		t.Error("got nil for /data/latency/tag protobuf")
	}

	if len(tr.Result) != 3 {
		t.Errorf("expected 3 tag results got %d", len(tr.Result))
	}

	if tr.Result[0].Tag != "DAGG" {
		t.Errorf("expected DAGG as the first tag got %s", tr.Result[0].Tag)
	}

	if tr.Result[0].SiteID != "TAUP" {
		t.Errorf("expected TAUP as the first siteID got %s", tr.Result[0].Tag)
	}

	if tr.Result[0].TypeID != "latency.strong" {
		t.Errorf("expected latency.stronge as the first typeID got %s", tr.Result[0].TypeID)
	}
}

// protobuf of data latency threshold
func TestDataLatencyThreshold(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	r := wt.Request{ID: wt.L(), URL: "/data/latency/threshold", Accept: "application/x-protobuf"}

	var b []byte
	var err error

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	var f mtrpb.DataLatencyThresholdResult

	if err = proto.Unmarshal(b, &f); err != nil {
		t.Error(err)
	}

	if f.Result == nil {
		t.Error("got nil for /data/latency/threshold protobuf")
	}

	if len(f.Result) != 1 {
		t.Error("expected 1 result.")
	}

	d := f.Result[0]

	if d.SiteID != "TAUP" {
		t.Errorf("expected TAUP got %s", d.SiteID)
	}

	if d.TypeID != "latency.strong" {
		t.Errorf("expected latency.strong got %s", d.TypeID)
	}

	if d.Upper != 15000 {
		t.Errorf("expected 15000 got %d", d.Upper)
	}

	if d.Lower != 12000 {
		t.Errorf("expected 12000 got %d", d.Lower)
	}
}
