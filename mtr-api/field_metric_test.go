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

// adds some test field metrics.  Assumes the caller has already called
//  setup(t)
//  defer teardown()
func addFieldMetrics(t *testing.T) {
	// Device model

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
	doRequest("PUT", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&time="+now.Truncate(time.Minute).Format(time.RFC3339)+"&value=10000", 200, t)
	doRequest("PUT", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&time="+now.Truncate(time.Minute).Format(time.RFC3339)+"&value=14100", 429, t)

}

func TestFieldMetrics(t *testing.T) {
	setup(t)
	defer teardown()

	addFieldMetrics(t)

	// Thresholds

	// Create a threshold on a metric
	doRequest("PUT", "*/*", "/field/metric/threshold?deviceID=gps-taupoairport&typeID=voltage&lower=12000&upper=15000", 200, t)

	// Update a threshold on a metric
	doRequest("PUT", "*/*", "/field/metric/threshold?deviceID=gps-taupoairport&typeID=voltage&lower=13000&upper=15000", 200, t)

	// Delete a threshold on a metric then create it again
	doRequest("DELETE", "*/*", "/field/metric/threshold?deviceID=gps-taupoairport&typeID=voltage", 200, t)
	doRequest("PUT", "*/*", "/field/metric/threshold?deviceID=gps-taupoairport&typeID=voltage&lower=12000&upper=15000", 200, t)

	// Tags

	// Create a tag on a metric type.  Multiple tags per metric are possible.  Repeat PUT is ok.
	doRequest("PUT", "*/*", "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=TAUP", 200, t)
	doRequest("PUT", "*/*", "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", 200, t)

	// Delete a tag on a metric
	doRequest("DELETE", "*/*", "/field/metric/tag?deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", 200, t)

	// GET requests
	// Non specific Accept headers return svg.

	// Model
	doRequest("GET", "application/json;version=1", "/field/model", 200, t)

	// Device
	doRequest("GET", "application/json;version=1", "/field/device", 200, t)

	// Metrics.  Resolution is optional on plots and sparks.  yrange is also optional.  If not set autoranges on the data.
	// Options for the plot parameter:
	// default = line plot.
	// line
	// scatter
	// spark (line)
	// spark-line
	// spark-scatter
	//
	// if yrange is not set then the yaxis autoranges between 0 and ymax.
	doRequest("GET", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage", 200, t)
	doRequest("GET", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=minute", 200, t)
	doRequest("GET", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=five_minutes", 200, t)
	doRequest("GET", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&resolution=hour", 200, t)
	doRequest("GET", "*/*", "/field/metric?deviceID=gps-taupoairport&typeID=voltage&yrange=0.0,25.0", 200, t)
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
	// doRequest("GET", "*/*", "/field/metric/summary?bbox=WhiteIsland&width=800typeID=voltage", 200, t)
	// doRequest("GET", "*/*", "/field/metric/summary?bbox=NewZealand&width=800&typeID=voltage", 200, t) // SVG medium size map.

	// All latest metrics as a FieldMetricLatestResult protobuf
	doRequest("GET", "application/x-protobuf", "/field/metric/summary", 200, t)
	// Latest voltage metrics
	doRequest("GET", "application/x-protobuf", "/field/metric/summary?typeID=voltage", 200, t)

	// Thresholds
	doRequest("GET", "application/json;version=1", "/field/metric/threshold", 200, t)

	// Tags
	doRequest("GET", "application/json;version=1", "/field/metric/tag", 200, t)          // All tags on metrics
	doRequest("GET", "application/json;version=1", "/field/metric/tag?tag=LINZ", 200, t) // All metrics for a tag
	doRequest("GET", "application/json;version=1", "/field/tag", 200, t)                 // All tag names no metrics

	// Metric types
	doRequest("GET", "application/json;version=1", "/field/type", 200, t) // All metrics type
}

func TestFieldMetricsSummary(t *testing.T) {
	setup(t)
	defer teardown()

	addFieldMetrics(t)

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

	if r.Upper != 0 {
		t.Errorf("expected 0 got %d", r.Upper)
	}

	if r.Lower != 0 {
		t.Errorf("expected 0 got %d", r.Lower)
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
	request.SetBasicAuth(userW, keyW)
	request.Header.Add("Accept", accept)

	if response, err = client.Do(request); err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if status != response.StatusCode {
		t.Errorf("Wrong response code for %s got %d expected %d", l, response.StatusCode, status)
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
