package main

// This file is auto generated - do not edit.
// It was created with weftgenapi from github.com/GeoNet/weft/weftgenapi

import (
	"bytes"
	"github.com/GeoNet/weft"
	"io/ioutil"
	"net/http"
)

var mux = http.NewServeMux()

func init() {
	mux.HandleFunc("/api-docs", weft.MakeHandlerPage(docHandler))
	mux.HandleFunc("/app", weft.MakeHandlerAPI(appHandler))
	mux.HandleFunc("/app/metric", weft.MakeHandlerAPI(appmetricHandler))
	mux.HandleFunc("/application/counter", weft.MakeHandlerAPI(applicationcounterHandler))
	mux.HandleFunc("/application/metric", weft.MakeHandlerAPI(applicationmetricHandler))
	mux.HandleFunc("/application/timer", weft.MakeHandlerAPI(applicationtimerHandler))
	mux.HandleFunc("/data/completeness", weft.MakeHandlerAPI(datacompletenessHandler))
	mux.HandleFunc("/data/completeness/summary", weft.MakeHandlerAPI(datacompletenesssummaryHandler))
	mux.HandleFunc("/data/completeness/tag", weft.MakeHandlerAPI(datacompletenesstagHandler))
	mux.HandleFunc("/data/completeness/type", weft.MakeHandlerAPI(datacompletenesstypeHandler))
	mux.HandleFunc("/data/latency", weft.MakeHandlerAPI(datalatencyHandler))
	mux.HandleFunc("/data/latency/summary", weft.MakeHandlerAPI(datalatencysummaryHandler))
	mux.HandleFunc("/data/latency/tag", weft.MakeHandlerAPI(datalatencytagHandler))
	mux.HandleFunc("/data/latency/threshold", weft.MakeHandlerAPI(datalatencythresholdHandler))
	mux.HandleFunc("/data/site", weft.MakeHandlerAPI(datasiteHandler))
	mux.HandleFunc("/data/type", weft.MakeHandlerAPI(datatypeHandler))
	mux.HandleFunc("/field/device", weft.MakeHandlerAPI(fielddeviceHandler))
	mux.HandleFunc("/field/metric", weft.MakeHandlerAPI(fieldmetricHandler))
	mux.HandleFunc("/field/metric/summary", weft.MakeHandlerAPI(fieldmetricsummaryHandler))
	mux.HandleFunc("/field/metric/tag", weft.MakeHandlerAPI(fieldmetrictagHandler))
	mux.HandleFunc("/field/metric/threshold", weft.MakeHandlerAPI(fieldmetricthresholdHandler))
	mux.HandleFunc("/field/model", weft.MakeHandlerAPI(fieldmodelHandler))
	mux.HandleFunc("/field/state", weft.MakeHandlerAPI(fieldstateHandler))
	mux.HandleFunc("/field/state/tag", weft.MakeHandlerAPI(fieldstatetagHandler))
	mux.HandleFunc("/field/type", weft.MakeHandlerAPI(fieldtypeHandler))
	mux.HandleFunc("/tag", weft.MakeHandlerAPI(tagHandler))
	mux.HandleFunc("/tag/", weft.MakeHandlerAPI(tagsHandler))
}

func docHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		by, err := ioutil.ReadFile("assets/api-docs/index.html")
		if err != nil {
			return weft.InternalServerError(err)
		}
		b.Write(by)
		return &weft.StatusOK
	default:
		return &weft.MethodNotAllowed
	}
}
func appHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return appIdProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	default:
		return &weft.MethodNotAllowed
	}
}

func appmetricHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "image/svg+xml":
			if res := weft.CheckQuery(r, []string{"applicationID", "group"}, []string{"resolution", "sourceID", "yrange"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return appMetricSvg(r, h, b)
		case "text/csv":
			if res := weft.CheckQuery(r, []string{"applicationID", "group"}, []string{"endDate", "resolution", "sourceID", "startDate"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "text/csv")
			return appMetricCsv(r, h, b)
		default:
			if res := weft.CheckQuery(r, []string{"applicationID", "group"}, []string{"resolution", "sourceID", "yrange"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return appMetricSvg(r, h, b)
		}
	default:
		return &weft.MethodNotAllowed
	}
}

func applicationcounterHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "PUT":
		if res := weft.CheckQuery(r, []string{"applicationID", "count", "instanceID", "time", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return applicationCounterPut(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func applicationmetricHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "PUT":
		if res := weft.CheckQuery(r, []string{"applicationID", "instanceID", "time", "typeID", "value"}, []string{}); !res.Ok {
			return res
		}
		return applicationMetricPut(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func applicationtimerHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "PUT":
		if res := weft.CheckQuery(r, []string{"applicationID", "average", "count", "fifty", "instanceID", "ninety", "sourceID", "time"}, []string{}); !res.Ok {
			return res
		}
		return applicationTimerPut(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func datacompletenessHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "image/svg+xml":
			if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{"plot", "resolution", "yrange"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return dataCompletenessSvg(r, h, b)
		default:
			if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{"plot", "resolution", "yrange"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return dataCompletenessSvg(r, h, b)
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{"count", "siteID", "time", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return dataCompletenessPut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return dataCompletenessDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func datacompletenesssummaryHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "image/svg+xml":
			if res := weft.CheckQuery(r, []string{"bbox", "typeID", "width"}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return dataCompletenessSummarySvg(r, h, b)
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{"typeID"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return dataCompletenessSummaryProto(r, h, b)
		default:
			if res := weft.CheckQuery(r, []string{"bbox", "typeID", "width"}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return dataCompletenessSummarySvg(r, h, b)
		}
	default:
		return &weft.MethodNotAllowed
	}
}

func datacompletenesstagHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return dataCompletenessTagProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{"siteID", "tag", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return dataCompletenessTagPut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{"siteID", "tag", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return dataCompletenessTagDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func datacompletenesstypeHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return dataCompletenessTypeProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	default:
		return &weft.MethodNotAllowed
	}
}

func datalatencyHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "image/svg+xml":
			if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{"plot", "resolution", "yrange"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return dataLatencySvg(r, h, b)
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{"resolution"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return dataLatencyProto(r, h, b)
		case "text/csv":
			if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{"endDate", "resolution", "startDate"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "text/csv")
			return dataLatencyCsv(r, h, b)
		default:
			if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{"plot", "resolution", "yrange"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return dataLatencySvg(r, h, b)
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{"mean", "siteID", "time", "typeID"}, []string{"fifty", "max", "min", "ninety"}); !res.Ok {
			return res
		}
		return dataLatencyPut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return dataLatencyDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func datalatencysummaryHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "image/svg+xml":
			if res := weft.CheckQuery(r, []string{"bbox", "typeID", "width"}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return dataLatencySummarySvg(r, h, b)
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{"typeID"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return dataLatencySummaryProto(r, h, b)
		default:
			if res := weft.CheckQuery(r, []string{"bbox", "typeID", "width"}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return dataLatencySummarySvg(r, h, b)
		}
	default:
		return &weft.MethodNotAllowed
	}
}

func datalatencytagHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{"siteID", "typeID"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return dataLatencyTagProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{"siteID", "tag", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return dataLatencyTagPut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{"siteID", "tag", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return dataLatencyTagDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func datalatencythresholdHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{"siteID", "typeID"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return dataLatencyThresholdProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{"lower", "siteID", "typeID", "upper"}, []string{}); !res.Ok {
			return res
		}
		return dataLatencyThresholdPut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{"siteID", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return dataLatencyThresholdDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func datasiteHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return dataSiteProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{"latitude", "longitude", "siteID"}, []string{}); !res.Ok {
			return res
		}
		return dataSitePut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{"siteID"}, []string{}); !res.Ok {
			return res
		}
		return dataSiteDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func datatypeHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return dataTypeProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	default:
		return &weft.MethodNotAllowed
	}
}

func fielddeviceHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return fieldDeviceProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{"deviceID", "latitude", "longitude", "modelID"}, []string{}); !res.Ok {
			return res
		}
		return fieldDevicePut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{"deviceID"}, []string{}); !res.Ok {
			return res
		}
		return fieldDeviceDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func fieldmetricHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{"deviceID", "typeID"}, []string{"resolution"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return fieldMetricProto(r, h, b)
		case "image/svg+xml":
			if res := weft.CheckQuery(r, []string{"deviceID", "typeID"}, []string{"plot", "resolution"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return fieldMetricSvg(r, h, b)
		case "text/csv":
			if res := weft.CheckQuery(r, []string{"deviceID", "typeID"}, []string{"endDate", "resolution", "startDate"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "text/csv")
			return fieldMetricCsv(r, h, b)
		default:
			if res := weft.CheckQuery(r, []string{"deviceID", "typeID"}, []string{"plot", "resolution"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return fieldMetricSvg(r, h, b)
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{"deviceID", "time", "typeID", "value"}, []string{}); !res.Ok {
			return res
		}
		return fieldMetricPut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{"deviceID", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return fieldMetricDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func fieldmetricsummaryHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{"typeID"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return fieldLatestProto(r, h, b)
		case "image/svg+xml":
			if res := weft.CheckQuery(r, []string{"bbox", "typeID", "width"}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return fieldLatestSvg(r, h, b)
		case "application/vnd.geo+json":
			if res := weft.CheckQuery(r, []string{"typeID"}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/vnd.geo+json")
			return fieldLatestGeoJSON(r, h, b)
		default:
			if res := weft.CheckQuery(r, []string{"bbox", "typeID", "width"}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "image/svg+xml")
			return fieldLatestSvg(r, h, b)
		}
	default:
		return &weft.MethodNotAllowed
	}
}

func fieldmetrictagHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{"deviceID", "typeID"}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return fieldMetricTagProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{"deviceID", "tag", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return fieldMetricTagPut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{"deviceID", "tag", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return fieldMetricTagDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func fieldmetricthresholdHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return fieldThresholdProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{"deviceID", "lower", "typeID", "upper"}, []string{}); !res.Ok {
			return res
		}
		return fieldThresholdPut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{"deviceID", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return fieldThresholdDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func fieldmodelHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return fieldModelProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{"modelID"}, []string{}); !res.Ok {
			return res
		}
		return fieldModelPut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{"modelID"}, []string{}); !res.Ok {
			return res
		}
		return fieldModelDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func fieldstateHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return fieldStateProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{"deviceID", "time", "typeID", "value"}, []string{}); !res.Ok {
			return res
		}
		return fieldStatePut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{"deviceID", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return fieldStateDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func fieldstatetagHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return fieldStateTagProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{"deviceID", "tag", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return fieldStateTagPut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{"deviceID", "tag", "typeID"}, []string{}); !res.Ok {
			return res
		}
		return fieldStateTagDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}

func fieldtypeHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return fieldTypeProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	default:
		return &weft.MethodNotAllowed
	}
}

func tagHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return tagsProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	default:
		return &weft.MethodNotAllowed
	}
}

func tagsHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/x-protobuf":
			if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
				return res
			}
			h.Set("Content-Type", "application/x-protobuf")
			return tagProto(r, h, b)
		default:
			return &weft.NotAcceptable
		}
	case "PUT":
		if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
			return res
		}
		return tagPut(r, h, b)
	case "DELETE":
		if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
			return res
		}
		return tagDelete(r, h, b)
	default:
		return &weft.MethodNotAllowed
	}
}
