package main

import "testing"

func addDataMetrics(t *testing.T) {

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
}

func TestDataMetrics(t *testing.T) {
	setup(t)
	defer teardown()

	addDataMetrics(t)
}
