package main

// The server for the mtr-ui, starting as simple as possible.
// TODO: keep 12 factor app principles in mind, use logentries, env vars instead of json config, etc.

import (
	_ "github.com/GeoNet/log/logentries"
	"log"
	"net/http"
	"net/url"
	"os"
)

var (
	mux           *http.ServeMux
	mtrApiUrl     *url.URL
	webServerPort string
)

func init() {
	var err error
	webServerPort = os.Getenv("MTR_UI_PORT")
	mtrApiUrlString := os.Getenv("MTR_API_URL")
	switch "" {
	case webServerPort:
		log.Fatal("error, environment variable MTR_UI_PORT must be set (eg: 8080)")
	case mtrApiUrlString:
		log.Fatal("error, environment variable MTR_API_URL must be set (eg: https://mtr-api.geonet.org.nz)")
	}

	if mtrApiUrl, err = url.Parse(mtrApiUrlString); err != nil {
		log.Fatal(err)
	}

	mux = http.NewServeMux()
	mux.HandleFunc("/", toHandler(homepageHandler))
	mux.HandleFunc("/field", toHandler(fieldPageHandler))
	mux.HandleFunc("/field/metrics", toHandler(fieldMetricsPageHandler))
	mux.HandleFunc("/field/devices", toHandler(fieldDevicesPageHandler))
	mux.HandleFunc("/field/plot", toHandler(fieldPlotPageHandler))
	mux.HandleFunc("/data", toHandler(dataPageHandler))
	mux.HandleFunc("/data/sites", toHandler(dataSitesPageHandler))
	mux.HandleFunc("/data/metrics", toHandler(dataMetricsPageHandler))
	mux.HandleFunc("/data/plot", toHandler(dataPlotPageHandler))
	mux.HandleFunc("/map/", toHandler(mapPageHandler))
	mux.HandleFunc("/field/metric", toHandler(metricDetailHandler))
	mux.HandleFunc("/search", toHandler(searchHandler))
}

func main() {
	log.Println("starting mtr-ui server")
	log.Fatal(http.ListenAndServe(":"+webServerPort, mux))
}
