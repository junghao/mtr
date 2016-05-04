package main

import (
	"testing"
	"time"
	"fmt"
)

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

	// Load some latency metrics (every 5 mins)
	now := time.Now().UTC()
	v := 14000
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

}

func TestDataMetrics(t *testing.T) {
	setup(t)
	defer teardown()

	addDataMetrics(t)

	// Add another site, some latency data, then delete.
	doRequest("DELETE", "*/*", "/data/site?siteID=WGTN", 200, t)
	doRequest("PUT", "*/*", "/data/site?siteID=WGTN&latitude=-38.74270&longitude=176.08100", 200, t)

	// min, max, fifty, ninety are optional latency values
	doRequest("PUT", "*/*", "/data/latency?siteID=WGTN&typeID=latency.strong&time="+time.Now().UTC().Format(time.RFC3339)+
	"&mean=10000&min=10&max=100000&fifty=9000&ninety=12000", 200, t)

	doRequest("DELETE", "*/*", "/data/latency?siteID=WGTN&typeID=latency.strong", 200, t)

	// Latency plots.  Resolution is optional on plots and sparks.  yrange is also optional.  If not set autoranges on the data.
	// Options for the plot parameter:
	// default = line plot.
	// line
	// scatter
	// spark (line)
	// spark-line
	// spark-scatter
	//
	// if yrange is not set then the yaxis autoranges between 0 and ymax.
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong", 200, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=minute", 200, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=five_minutes", 200, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=hour", 200, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&yrange=0.0,25.0", 200, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=minute", 200, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=hour", 200, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&resolution=day", 400, t)
	doRequest("GET", "*/*", "/data/latency?siteID=TAUP&typeID=latency.strong&plot=spark", 200, t)
}
