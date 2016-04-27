package main

// The server for the mtr-ui, starting as simple as possible.
// TODO: keep 12 factor app principles in mind, use logentries, env vars instead of json config, etc.

import (
	_ "github.com/GeoNet/log/logentries"
	"log"
	"net/http"
	"os"
)

var (
	mux *http.ServeMux
	webServerPort string
)

func init() {
	webServerPort = os.Getenv("MTR_UI_PORT")
	switch "" {
	case webServerPort:
		log.Fatal("error, environment variable MTR_UI_PORT must be set (eg: 8080)")
	}

	mux = http.NewServeMux()
	mux.HandleFunc("/", toHandler(handler))
}

func main() {
	log.Println("starting mtr-ui server")
	log.Fatal(http.ListenAndServe(":"+webServerPort, mux))
}
