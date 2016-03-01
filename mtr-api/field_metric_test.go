package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

var client = &http.Client{}

func TestFieldMetrics(t *testing.T) {
	setup(t)
	defer teardown()

	// Localities

	// Deleting a locality also deletes any metrics for the locality.
	doRequest("DELETE", "*/*", "/field/locality?localityID=taupoairport", 200, t)
	doRequest("DELETE", "*/*", "/field/locality?localityID=seatounschoolwellington", 200, t)
	doRequest("DELETE", "*/*", "/field/locality?localityID=temaari", 200, t)

	// Creates a locality.  Repeated requests update the non localityID parameters
	doRequest("PUT", "*/*", "/field/locality?localityID=taupoairport&name=Taupo+Airport&latitude=-38.74270&longitude=176.08100", 200, t)
	doRequest("PUT", "*/*", "/field/locality?localityID=seatounschoolwellington&name=Seatoun+School+Wellington&latitude=-41.32645&longitude=174.83764", 200, t)
	doRequest("PUT", "*/*", "/field/locality?localityID=temaari&name=Te+Maari&latitude=-39.11561&longitude=175.70406", 200, t)

	// Device model

	// Creates a device model.  Repeated requests noop.
	doRequest("PUT", "*/*", "/field/model?modelID=Trimble+NetR9", 200, t)

	// Delete a model (then recreate it).  Delete cascades to devices with that model (which cascades to metrics)
	doRequest("DELETE", "*/*", "/field/model?modelID=Trimble+NetR9", 200, t)
	doRequest("PUT", "*/*", "/field/model?modelID=Trimble+NetR9", 200, t)

	// Devices
	doRequest("PUT", "*/*", "/field/device?deviceID=gps-taupoairport&modelID=Trimble+NetR9", 200, t)
	doRequest("DELETE", "*/*", "/field/device?deviceID=gps-taupoairport", 200, t)
	doRequest("PUT", "*/*", "/field/device?deviceID=gps-taupoairport&modelID=Trimble+NetR9", 200, t)

	// Metrics

	// Load some metrics (every 5 mins)
	now := time.Now().UTC()
	v := 14000
	for i := -1440; i < 0; i += 5 {
		if i >= -100 {
			v = int(14000*(1/(float64(i)+101.0))) + 10000
			if v > 14000 {
				v = 14000
			}
		}

		doRequest("PUT", "*/*", fmt.Sprintf("/field/metric?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&time=%s&value=%d",
			now.Add(time.Duration(i)*time.Minute).Format(time.RFC3339), v), 200, t)
		doRequest("PUT", "*/*", fmt.Sprintf("/field/metric?localityID=seatounschoolwellington&deviceID=gps-taupoairport&typeID=voltage&time=%s&value=%d",
			now.Add(time.Duration(i)*time.Minute).Format(time.RFC3339), v), 200, t)
		doRequest("PUT", "*/*", fmt.Sprintf("/field/metric?localityID=seatounschoolwellington&deviceID=gps-taupoairport&typeID=satellites&time=%s&value=%d",
			now.Add(time.Duration(i)*time.Minute).Format(time.RFC3339), v), 200, t)
	}

	// It is not an error to send more than one metric in an minute.  The last value sent is saved
	doRequest("PUT", "*/*", "/field/metric?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&time="+now.Truncate(time.Minute).Format(time.RFC3339)+"&value=10000", 200, t)
	doRequest("PUT", "*/*", "/field/metric?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&time="+now.Truncate(time.Minute).Format(time.RFC3339)+"&value=14100", 200, t)

	// Delete all metrics for a device type at a locality
	doRequest("DELETE", "*/*", "/field/metric?localityID=seatounschoolwellington&deviceID=gps-taupoairport&typeID=satellites", 200, t)

	// Thresholds

	// Create a threshold on a metric
	doRequest("PUT", "*/*", "/field/metric/threshold?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&lower=12000&upper=15000", 200, t)

	// Update a threshold on a metric
	doRequest("PUT", "*/*", "/field/metric/threshold?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&lower=13000&upper=15000", 200, t)

	// Delete a threshold on a metric then create it again
	doRequest("DELETE", "*/*", "/field/metric/threshold?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage", 200, t)
	doRequest("PUT", "*/*", "/field/metric/threshold?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&lower=12000&upper=15000", 200, t)

	// Tags

	// Create a tag on a metric type.  Multiple tags per metric are possible.  Repeat PUT is ok.
	doRequest("PUT", "*/*", "/field/metric/tag?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&tag=TAUP", 200, t)
	doRequest("PUT", "*/*", "/field/metric/tag?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", 200, t)

	// Delete a tag on a metric
	doRequest("DELETE", "*/*", "/field/metric/tag?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&tag=LINZ", 200, t)

	// GET requests
	// Non specific Accept headers return svg.

	// All localities
	doRequest("GET", "application/vnd.geo+json;version=1", "/field/locality", 200, t) // SVG icon map
	doRequest("GET", "application/json;version=1", "/field/locality", 200, t)

	// Single locality
	doRequest("GET", "*/*", "/field/locality?localityID=taupoairport", 200, t)
	doRequest("GET", "application/json;version=1", "/field/locality?localityID=taupoairport", 200, t)
	doRequest("GET", "application/vnd.geo+json;version=1", "/field/locality?localityID=taupoairport", 200, t)

	// Non exisent locality
	doRequest("GET", "*/*", "/field/locality?localityID=nope", 404, t)
	doRequest("GET", "application/json;version=1", "/field/locality?localityID=nope", 404, t)
	doRequest("GET", "application/vnd.geo+json;version=1", "/field/locality?localityID=nope", 404, t)

	// Dark localities - no metrics.
	doRequest("GET", "*/*", "/field/locality/dark", 200, t)
	doRequest("GET", "application/json;version=1", "/field/locality/dark", 200, t)
	doRequest("GET", "application/vnd.geo+json;version=1", "/field/locality/dark", 200, t)

	// Model
	doRequest("GET", "application/json;version=1", "/field/model", 200, t)

	// Device
	doRequest("GET", "application/json;version=1", "/field/device", 200, t)

	// Metrics.  Resolution is optional on plots and sparks.  yrange is also optional.  If not set autoranges on the data.
	doRequest("GET", "*/*", "/field/metric?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage", 200, t)
	doRequest("GET", "*/*", "/field/metric?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&yrange=0.0,25.0", 200, t)
	doRequest("GET", "*/*", "/field/metric?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&resolution=minute", 200, t)
	doRequest("GET", "*/*", "/field/metric?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&resolution=hour", 200, t)
	doRequest("GET", "*/*", "/field/metric?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&resolution=day", 200, t)
	doRequest("GET", "*/*", "/field/metric?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage&plot=spark", 200, t)
	doRequest("GET", "text/csv", "/field/metric?localityID=taupoairport&deviceID=gps-taupoairport&typeID=voltage", 200, t)

	// Latest metrics as SVG map
	//  These only pass with the map180 data in the DB.
	// Values for bbox and insetBbox are ChathamIsland LakeTaupo NewZealand NewZealandRegion
	// RaoulIsland WhiteIsland
	// or lhe bounding box for the map defining the lower left and upper right longitude
	// latitude (EPSG:4327) corners e.g., <code>165,-48,179,-34</code>.  Latitude must be in the range -85 to 85.  Maps can be 180 centric and bbox
	// definitions for longitude can be -180 to 180 or 0 to 360
	//
	// doRequest("GET", "*/*", "/field/metric/latest?bbox=NewZealand&width=800", 200, t)
	// doRequest("GET", "*/*", "/field/metric/latest?bbox=WhiteIsland&width=800&insetBbox=NewZealand", 200, t)
	// doRequest("GET", "*/*", "/field/metric/latest?bbox=NewZealand&width=800&typeID=voltage", 200, t) // SVG medium size map.

	doRequest("GET", "application/json;version=1", "/field/metric/latest", 200, t)
	doRequest("GET", "application/vnd.geo+json;version=1", "/field/metric/latest", 200, t)

	doRequest("GET", "application/json;version=1", "/field/metric/latest?typeID=voltage", 200, t)
	doRequest("GET", "application/vnd.geo+json;version=1", "/field/metric/latest?typeID=voltage", 200, t)

	// Thresholds
	doRequest("GET", "application/json;version=1", "/field/metric/threshold", 200, t)

	// Tags
	doRequest("GET", "application/json;version=1", "/field/metric/tag", 200, t)          // All tags
	doRequest("GET", "application/json;version=1", "/field/metric/tag?tag=nope", 404, t) // tag that doesn't exist.

	// Metric types
	doRequest("GET", "application/json;version=1", "/field/metric/type", 200, t) // All metrics type
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
