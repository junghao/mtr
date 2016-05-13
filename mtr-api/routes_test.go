package main

import (
	"fmt"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

var client = &http.Client{}

func TestRoutes(t *testing.T) {
	setup(t)
	defer teardown()

	/*
		Field metrics
	*/

	// Creates a device model.  Repeated requests noop.
	doRequest("PUT", "*/*", "/field/model?modelID=Trimble+NetR9", 200, t)

	// Delete a model (then recreate it).  Delete cascades to devices with that model (which cascades to metrics)
	doRequest("DELETE", "*/*", "/field/model?modelID=Trimble+NetR9", 200, t)
	doRequest("PUT", "*/*", "/field/model?modelID=Trimble+NetR9", 200, t)

	// Devices are at a lat long
	doRequest("PUT", "*/*", "/field/device?deviceID=gps-taupoairport&modelID=Trimble+NetR9&latitude=-38.74270&longitude=176.08100", 200, t)
	doRequest("DELETE", "*/*", "/field/device?deviceID=gps-taupoairport", 200, t)
	doRequest("PUT", "*/*", "/field/device?deviceID=gps-taupoairport&modelID=Trimble+NetR9&latitude=-38.74270&longitude=176.08100", 200, t)

	doRequest("DELETE", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage", 200, t)

	// Load some metrics (every 5 mins)
	now := time.Now().UTC()
	v := 14000
	for i := -720; i < 0; i += 5 {
		if i >= -100 {
			v = int(14000*(1/(float64(i)+101.0))) + 10000
			if v > 14000 {
				v = 14000
			}
		}

		doRequest("PUT", "*/*", fmt.Sprintf("/field/metric?deviceID=gps-taupoairport&typeID=voltage&time=%s&value=%d",
			now.Add(time.Duration(i)*time.Minute).Format(time.RFC3339), v), 200, t)
	}

	// Should get a rate limit error for sends in the same minute
	doRequest("PUT", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&time="+now.Truncate(time.Minute).Format(time.RFC3339)+"&value=14100", 200, t)
	doRequest("PUT", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&time="+now.Truncate(time.Minute).Format(time.RFC3339)+"&value=15100", 429, t)

	// Tags

	doRequest("DELETE", "*/*", "/tag/LINZ", 200, t)

	// tag must exist before it can be added to a metric
	doRequest("PUT", "*/*", "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", 400, t)

	doRequest("PUT", "*/*", "/tag/LINZ", 200, t)
	doRequest("PUT", "*/*", "/tag/TAUP", 200, t)

	// Create a tag on a metric type.  Multiple tags per metric are possible.  Repeat PUT is ok.
	doRequest("PUT", "*/*", "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=TAUP", 200, t)
	doRequest("PUT", "*/*", "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", 200, t)

	// Delete a tag on a metric
	doRequest("DELETE", "*/*", "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", 200, t)

	if _, err := db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY field.metric_summary`); err != nil {
		t.Error(err)
	}

	// Thresholds

	// Create a threshold on a metric
	doRequest("PUT", "*/*", "/field/metric/threshold?deviceID=gps-taupoairport&typeID=voltage&lower=12000&upper=15000", 200, t)

	// Update a threshold on a metric
	doRequest("PUT", "*/*", "/field/metric/threshold?deviceID=gps-taupoairport&typeID=voltage&lower=13000&upper=15000", 200, t)

	// Delete a threshold on a metric then create it again
	doRequest("DELETE", "*/*", "/field/metric/threshold?deviceID=gps-taupoairport&typeID=voltage", 200, t)
	doRequest("PUT", "*/*", "/field/metric/threshold?deviceID=gps-taupoairport&typeID=voltage&lower=12000&upper=45000", 200, t)

	if _, err := db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY field.metric_summary`); err != nil {
		t.Error(err)
	}

	// GET requests
	// Non specific Accept headers return svg.

	// Model
	doRequest("GET", "application/json;version=1", "/field/model", 200, t)

	// Device
	doRequest("GET", "application/json;version=1", "/field/device", 200, t)

	// Metrics.  Resolution is optional on plots.  Resolution is fixed for sparks.
	// Options for the plot parameter:
	// line [default] = line plot.
	// spark = spark line
	doRequest("GET", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage", 200, t)
	doRequest("GET", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=minute", 200, t)
	doRequest("GET", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=five_minutes", 200, t)
	doRequest("GET", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=hour", 200, t)
	doRequest("GET", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=minute", 200, t)
	doRequest("GET", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=hour", 200, t)
	doRequest("GET", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=day", 400, t)
	doRequest("GET", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&plot=spark", 200, t)

	// Latest metrics as SVG map
	//  These only pass with the map180 data in the DB.
	// Values for bbox and insetBbox are ChathamIsland LakeTaupo NewZealand NewZealandRegion
	// RaoulIsland WhiteIsland
	// or lhe bounding box for the map defining the lower left and upper right longitude
	// latitude (EPSG:4327) corners e.g., <code>165,-48,179,-34</code>.  Latitude must be in the range -85 to 85.  Maps can be 180 centric and bbox
	// definitions for longitude can be -180 to 180 or 0 to 360
	//
	//doRequest("GET", "*/*", "/field/metric/summary?bbox=WhiteIsland&width=800&typeID=voltage", 200, t)
	//doRequest("GET", "*/*", "/field/metric/summary?bbox=NewZealand&width=800&typeID=voltage", 200, t)

	// All latest metrics as a FieldMetricLatestResult protobuf
	doRequest("GET", "application/x-protobuf", "/field/metric/summary", 200, t)
	// Latest voltage metrics
	doRequest("GET", "application/x-protobuf", "/field/metric/summary?typeID=voltage", 200, t)

	// Thresholds
	doRequest("GET", "application/json;version=1", "/field/metric/threshold", 200, t)

	// Metric types
	doRequest("GET", "application/json;version=1", "/field/type", 200, t) // All metrics type

	/*
		Data Latency
	*/

	// Delete site - cascades to metrics
	doRequest("DELETE", "*/*", "/data/site?siteID=TAUP", 200, t)

	// create a site.  Lat lon are indicative only and may not be suitable for
	// precise data use.
	doRequest("PUT", "*/*", "/data/site?siteID=TAUP&latitude=-38.74270&longitude=176.08100", 200, t)
	// update the site location
	doRequest("PUT", "*/*", "/data/site?siteID=TAUP&latitude=-38.64270&longitude=176.08100", 200, t)
	// delete then recreate
	doRequest("DELETE", "*/*", "/data/site?siteID=TAUP", 200, t)
	doRequest("PUT", "*/*", "/data/site?siteID=TAUP&latitude=-38.74270&longitude=176.08100", 200, t)

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

		doRequest("PUT", "*/*", fmt.Sprintf("/data/latency?siteID=TAUP&typeID=latency.strong&time=%s&mean=%d",
			now.Add(time.Duration(i)*time.Minute).Format(time.RFC3339), v), 200, t)
	}

	// Should get a rate limit error for sends in the same minute
	doRequest("PUT", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&time="+now.Truncate(time.Minute).Format(time.RFC3339)+"&mean=10000", 200, t)
	doRequest("PUT", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&time="+now.Truncate(time.Minute).Format(time.RFC3339)+"&mean=14100", 429, t)

	// Refresh the latency_summary view.  Usually done on timer in server.go
	if _, err := db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY data.latency_summary`); err != nil {
		t.Error(err)
	}

	// Add another site, some latency data, then delete.
	doRequest("DELETE", "*/*", "/data/site?siteID=WGTN", 200, t)
	doRequest("PUT", "*/*", "/data/site?siteID=WGTN&latitude=-38.74270&longitude=176.08100", 200, t)

	// min, max, fifty, ninety are optional latency values
	doRequest("PUT", "*/*", "/data/latency?siteID=WGTN&typeID=latency.strong&time="+time.Now().UTC().Format(time.RFC3339)+
		"&mean=10000&min=10&max=100000&fifty=9000&ninety=12000", 200, t)

	doRequest("DELETE", "*/*", "/data/latency?siteID=WGTN&typeID=latency.strong", 200, t)

	// Create a threshold for latency.
	// I assume a single threshold would be for mean, fifty, and ninety?
	doRequest("DELETE", "*/*", "/data/latency/threshold?siteID=TAUP&typeID=latency.strong", 200, t)
	doRequest("PUT", "*/*", "/data/latency/threshold?siteID=TAUP&typeID=latency.strong&lower=12000&upper=15000", 200, t)

	// Update a threshold
	doRequest("PUT", "*/*", "/data/latency/threshold?siteID=TAUP&typeID=latency.strong&lower=13000&upper=15000", 200, t)

	// Delete a threshold then create it again
	doRequest("DELETE", "*/*", "/data/latency/threshold?siteID=TAUP&typeID=latency.strong", 200, t)
	doRequest("PUT", "*/*", "/data/latency/threshold?siteID=TAUP&typeID=latency.strong&lower=12000&upper=15000", 200, t)

	// Tags

	doRequest("DELETE", "*/*", "/tag/FRED", 200, t)
	doRequest("DELETE", "*/*", "/tag/DAGG", 200, t)

	// tag must exist before it can be added to a metric
	doRequest("PUT", "*/*", "/data/latency/tag?siteID=FRED&typeID=latency.strong&tag=TAUP", 400, t)

	doRequest("PUT", "*/*", "/tag/FRED", 200, t)
	doRequest("PUT", "*/*", "/tag/DAGG", 200, t)
	doRequest("PUT", "*/*", "/tag/TAUP", 200, t)

	// Create a tag on a latency.  Multiple tags per metric are possible.  Repeat PUT is ok.
	doRequest("PUT", "*/*", "/data/latency/tag?siteID=TAUP&typeID=latency.strong&tag=FRED", 200, t)
	doRequest("PUT", "*/*", "/data/latency/tag?siteID=TAUP&typeID=latency.strong&tag=DAGG", 200, t)
	doRequest("PUT", "*/*", "/data/latency/tag?siteID=TAUP&typeID=latency.strong&tag=TAUP", 200, t)

	// Latency plots.  Resolution is optional on plots and sparks.
	// Options for the plot parameter:
	// line [default] = line plot.
	// spark = spark line.
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong", 200, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=minute", 200, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=five_minutes", 200, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=hour", 200, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=minute", 200, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=hour", 200, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=day", 400, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&plot=spark", 200, t)

	// Tags

	doRequest("DELETE", "*/*", "/tag/LINZ", 200, t)

	// tag must exist before it can be added to a metric
	doRequest("PUT", "*/*", "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", 400, t)

	doRequest("PUT", "*/*", "/tag/LINZ", 200, t)
	doRequest("PUT", "*/*", "/tag/TAUP", 200, t)

	// Create a tag on a metric type.  Multiple tags per metric are possible.  Repeat PUT is ok.
	doRequest("PUT", "*/*", "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=TAUP", 200, t)
	doRequest("PUT", "*/*", "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", 200, t)

	// Delete a tag on a metric
	doRequest("DELETE", "*/*", "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", 200, t)

	//doRequest("PUT", "*/*", "/tag/TAUP", 200, t)
	//doRequest("DELETE", "*/*", "/tag/TAUP", 200, t)
	//doRequest("PUT", "*/*", "/tag/TAUP", 200, t)

	if _, err := db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY data.latency_summary`); err != nil {
		t.Error(err)
	}

	if _, err := db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY field.metric_summary`); err != nil {
		t.Error(err)
	}

	// These tests use the data loaded above.
	testFieldMetricsSummary(t)
	testDataLatencySummary(t)
	testTagAllProto(t)
	testTagProto(t)
}

func testTagProto(t *testing.T) {

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

func testTagAllProto(t *testing.T) {

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

func testDataLatencySummary(t *testing.T) {

	doRequest("GET", "application/x-protobuf", "/data/latency/summary", 200, t)

	var err error
	var b []byte

	if b, err = getBytes("application/x-protobuf", "/data/latency/summary"); err != nil {
		t.Error(err)
	}

	var f mtrpb.DataLatencySummaryResult

	if err = proto.Unmarshal(b, &f); err != nil {
		t.Error(err)
	}

	if len(f.Result) != 1 {
		t.Error("expected 1 result.")
	}

	r := f.Result[0]

	if r.SiteID != "TAUP" {
		t.Errorf("expected TAUP got %s", r.SiteID)
	}

	if r.TypeID != "latency.strong" {
		t.Errorf("expected latency.strong got %s", r.TypeID)
	}

	if r.Mean != 10000 {
		t.Errorf("expected 10000 got %d", r.Mean)
	}

	if r.Fifty != 0 {
		t.Errorf("expected 0 got %d", r.Fifty)
	}

	if r.Ninety != 0 {
		t.Errorf("expected 0 got %d", r.Ninety)
	}

	if r.Seconds == 0 {
		t.Error("unexpected zero seconds")
	}

	if r.Upper != 15000 {
		t.Errorf("expected 15000 got %d", r.Upper)
	}

	if r.Lower != 12000 {
		t.Errorf("expected 12000 got %d", r.Lower)
	}

	doRequest("GET", "application/x-protobuf", "/data/latency/summary?typeID=latency.strong", 200, t)

	if b, err = getBytes("application/x-protobuf", "/data/latency/summary?typeID=latency.strong"); err != nil {
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

func testFieldMetricsSummary(t *testing.T) {
	doRequest("GET", "application/x-protobuf", "/field/metric/summary", 200, t)

	var err error
	var b []byte

	if b, err = getBytes("application/x-protobuf", "/field/metric/summary"); err != nil {
		t.Error(err)
	}

	var f mtrpb.FieldMetricSummaryResult

	if err = proto.Unmarshal(b, &f); err != nil {
		t.Error(err)
	}

	if len(f.Result) != 1 {
		t.Error("expected 1 result.")
	}

	r := f.Result[0]

	if r.DeviceID != "gps-taupoairport" {
		t.Errorf("expected gps-taupoairport got %s", r.DeviceID)
	}

	if r.ModelID != "Trimble NetR9" {
		t.Errorf("expected Trimble NetR9 got %s", r.ModelID)
	}

	if r.TypeID != "voltage" {
		t.Errorf("expected voltage got %s", r.TypeID)
	}

	if r.Value != 14100 {
		t.Errorf("expected 14100 got %d", r.Value)
	}

	if r.Seconds == 0 {
		t.Error("unexpected zero seconds")
	}

	if r.Upper != 45000 {
		t.Errorf("expected 45000 got %d", r.Upper)
	}

	if r.Lower != 12000 {
		t.Errorf("expected 12000 got %d", r.Lower)
	}

	// should be no errors and empty result for typeID=conn
	doRequest("GET", "application/x-protobuf", "/field/metric/summary?typeID=conn", 200, t)

	if b, err = getBytes("application/x-protobuf", "/field/metric/summary?typeID=conn"); err != nil {
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

func getBytes(accept, uri string) (b []byte, err error) {
	var request *http.Request
	var response *http.Response

	if request, err = http.NewRequest("GET", testServer.URL+uri, nil); err != nil {
		return
	}

	request.Header.Add("Accept", accept)

	if response, err = client.Do(request); err != nil {
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = fmt.Errorf("non 200 response: %d for %s", response.StatusCode, uri)
	}

	b, err = ioutil.ReadAll(response.Body)

	return
}

func doRequest(method, accept, uri string, status int, t *testing.T) {
	var request *http.Request
	var response *http.Response
	var err error
	l := loc()

	if request, err = http.NewRequest(method, testServer.URL+uri, nil); err != nil {
		t.Fatal(err)
	}

	request.Header.Add("Accept", accept)

	if method != "GET" {
		// Check that we have to be authenticated for non GET requests.
		// Run the all first with out auth
		var resp *http.Response
		if resp, err = client.Do(request); err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Wrong response code for %s with no auth should be status.Unauthorized", l)
		}

		request.SetBasicAuth(userW, keyW)
	}

	if response, err = client.Do(request); err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if status != response.StatusCode {
		t.Errorf("Wrong response code for %s got %d expected %d", l, response.StatusCode, status)
		by, _ := ioutil.ReadAll(response.Body)
		t.Log(string(by))
	}

	if method == "GET" && status == http.StatusOK {
		switch accept {
		case "*/*":
			if response.Header.Get("Content-Type") != "image/svg+xml" {
				t.Errorf("Wrong Content-Type for %s got %s expected image/svg+xml", l, response.Header.Get("Content-Type"))
			}
		default:
			if response.Header.Get("Content-Type") != accept {
				t.Errorf("Wrong Content-Type for %s got %s expected %s", l, response.Header.Get("Content-Type"), accept)
			}
		}
	}
}
