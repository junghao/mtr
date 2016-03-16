package mtrapp

import (
	"bytes"
	"encoding/json"
	"github.com/GeoNet/mtr/internal"
	"log"
	"net/http"
	"os"
	"time"
)

const timeout = 3 * time.Minute

var client = &http.Client{}
var send chan internal.AppMetrics

/*
SendMetrics takes Metics from send and sends them to the metrics server.
In the case of connection failures send retries with back off for up to three minutes
before discarding.
*/
func sendMetrics(send <-chan internal.AppMetrics) {
	var res *http.Response
	var req *http.Request
	var buf bytes.Buffer
	var err error

	req, err = http.NewRequest("POST", os.Getenv("MTR_SERVER")+"/app/metric", &buf)
	if err != nil {
		log.Println(err)
	}

	req.SetBasicAuth(os.Getenv("MTR_USER"), os.Getenv("MTR_KEY"))

	encoder := json.NewEncoder(&buf)

	for {
		select {
		case m := <-send:
			buf.Reset()
			err = encoder.Encode(m)
			if err != nil {
				log.Println(err)
				continue
			}

			deadline := time.Now().Add(timeout)

			for tries := 0; time.Now().Before(deadline); tries++ {
				if res, err = client.Do(req); err == nil {
					if res != nil && res.StatusCode != 200 {
						log.Printf("Non 200 code from metrics: %d", res.StatusCode)
					}
					break
				}
				log.Printf("server not responding (%s); backing off and retrying...", err)
				time.Sleep(time.Second << uint(tries))
			}
			if res != nil {
				res.Body.Close()
			}
		}
	}
}
