package main

import (
	wt "github.com/GeoNet/weft/wefttest"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
}

var testServer *httptest.Server

var routes = wt.Requests{
	// Top level pages
	{ID: wt.L(), URL: "/"},
	{ID: wt.L(), URL: "/data"},
	{ID: wt.L(), URL: "/field"},
	{ID: wt.L(), URL: "/map"},
	{ID: wt.L(), URL: "/tag"},

	// data pages
	{ID: wt.L(), URL: "/data/sites"},
	{ID: wt.L(), URL: "/data/plot?siteID=AUCK&typeID=latency.gnss.1hz"},
	{ID: wt.L(), URL: "/data/plot?siteID=CHTI&typeID=latency.gnss.1hz&resolution=minutes"},
	{ID: wt.L(), URL: "/data/plot?siteID=CHTI&typeID=latency.gnss.1hz&resolution=five_minutes"},
	{ID: wt.L(), URL: "/data/plot?siteID=CHTI&typeID=latency.gnss.1hz&resolution=hour"},
	{ID: wt.L(), URL: "/data/metrics"},
	{ID: wt.L(), URL: "/data/metrics?typeID=latency.gnss.1hz"},
	{ID: wt.L(), URL: "/data/metrics?typeID=latency.gnss.1hz&status=good"},
	{ID: wt.L(), URL: "/data/metrics?&status=good"},

	// field pages
	{ID: wt.L(), URL: "/field/"},
	{ID: wt.L(), URL: "/field/devices"},
	{ID: wt.L(), URL: "/field/devices?modelID=Bay%20City%20VSAT%20IDU"},
	{ID: wt.L(), URL: "/field/plot?deviceID=baycity-soundstage&typeID=ping"},
	{ID: wt.L(), URL: "/field/plot?deviceID=baycity-soundstage&typeID=ping&resolution=minutes"},
	{ID: wt.L(), URL: "/field/plot?deviceID=baycity-soundstage&typeID=ping&resolution=five_minutes"},
	{ID: wt.L(), URL: "/field/plot?deviceID=baycity-soundstage&typeID=ping&resolution=hour"},
	{ID: wt.L(), URL: "/field/metrics?modelID=Bay%20City%20VSAT%20IDU"},
	{ID: wt.L(), URL: "/field/devices?modelID=Bay%20City%20VSAT%20IDU&status=good"},
	{ID: wt.L(), URL: "/field/metrics?&status=good"},
	{ID: wt.L(), URL: "/field/metrics?typeID=centre&status=good"},

	// map pages
	{ID: wt.L(), URL: "/map/"},
	{ID: wt.L(), URL: "/map/conn"},
	{ID: wt.L(), URL: "/map/voltage"},
	{ID: wt.L(), URL: "/map/ping"},

	// applications page
	{ID: wt.L(), URL: "/app"},
	// this page uses the applicationID to construct img elements which are not validated:
	{ID: wt.L(), URL: "/app/plot?applicationID=test-app"},

	// tag page
	{ID: wt.L(), URL: "/tag/"},
	{ID: wt.L(), URL: "/tag/A-C"},

	// search
	{ID: wt.L(), URL: "/search?tagQuery=TAKP"},
	{ID: wt.L(), URL: "/search?tagQuery=TAKP&page=1"},
}

func setup() {
	testServer = httptest.NewServer(testHandler())
}
func teardown() {
	testServer.Close()
}
func testHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(w, r)
	})
}

// Test all routes give the expected response.  Also check with
// cache busters and extra query paramters.
func TestRoutes(t *testing.T) {
	setup()
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
