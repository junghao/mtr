package main

import (
	"bytes"
	"github.com/GeoNet/mtr/mtrapp"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type result struct {
	ok   bool   // set true to indicated success
	code int    // http status code for writing back the client e.g., http.StatusOK for success.
	msg  string // any error message for logging or to send to the client.
}

/*
requestHandler for handling http requests.  The response for the request
should be written into.  Any header values for the client can be set in h
e.g., Content-Type.
*/
type requestHandler func(r *http.Request, h http.Header, b *bytes.Buffer) *result

var (
	statusOK         = result{ok: true, code: http.StatusOK, msg: ""}
	methodNotAllowed = result{ok: false, code: http.StatusMethodNotAllowed, msg: "method not allowed"}
	notFound         = result{ok: false, code: http.StatusNotFound, msg: ""}
	notAcceptable    = result{ok: false, code: http.StatusNotAcceptable, msg: "specify accept"}
)

func internalServerError(err error) *result {
	return &result{ok: false, code: http.StatusInternalServerError, msg: err.Error()}
}

func badRequest(message string) *result {
	return &result{ok: false, code: http.StatusBadRequest, msg: message}
}

/*
checkQuery inspects r and makes sure all required query parameters
are present and that no more than the required and optional parameters
are present.
*/
func checkQuery(r *http.Request, required, optional []string) *result {
	if strings.Contains(r.URL.Path, ";") {
		return badRequest("cache buster")
	}

	v := r.URL.Query()

	if len(required) == 0 && len(optional) == 0 {
		if len(v) == 0 {
			return &statusOK
		} else {
			return badRequest("found unexpected query parameters")
		}
	}

	var missing []string

	for _, k := range required {
		if v.Get(k) == "" {
			missing = append(missing, k)
		} else {
			v.Del(k)
		}
	}

	switch len(missing) {
	case 0:
	case 1:
		return badRequest("missing required query parameter: " + missing[0])
	default:
		return badRequest("missing required query parameters: " + strings.Join(missing, ", "))
	}

	for _, k := range optional {
		v.Del(k)
	}

	if len(v) > 0 {
		return badRequest("found additional query parameters")
	}

	return &statusOK
}

/*
toHandler adds basic auth to f and returns a handler.
*/
func toHandler(f requestHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Find the name of the function f to use as the timer id
		id := r.Method
		fn := runtime.FuncForPC(reflect.ValueOf(f).Pointer())
		if fn != nil {
			id = fn.Name() + "." + id
		}

		t := mtrapp.Start(id)

		mtrapp.Requests.Inc()

		// TODO review return codes from handlers - missing things should 404, not 400.

		switch r.Method {
		case "PUT", "DELETE":
			if user, password, ok := r.BasicAuth(); ok && userW == user && keyW == password {
				// PUT and DELETE do not have a response body for the client so pass a nil buffer.
				res := f(r, w.Header(), nil)
				t.Stop()

				switch res.code {
				case http.StatusOK:
					w.WriteHeader(http.StatusOK)

					t.Track()
					if t.Taken() > 500 {
						log.Printf("%s took %d ms to handle %s", id, t.Taken(), r.URL.Path)
					}

					mtrapp.StatusOK.Inc()
				case http.StatusBadRequest:
					http.Error(w, res.msg, res.code)
					mtrapp.StatusBadRequest.Inc()
				case http.StatusInternalServerError:
					http.Error(w, res.msg, res.code)
					mtrapp.StatusInternalServerError.Inc()
					log.Printf("500 serving %s %s %s", r.Method, r.URL, res.msg)
				default:
					http.Error(w, res.msg, res.code)
				}

			} else {
				http.Error(w, "Access denied", http.StatusUnauthorized)
				mtrapp.StatusUnauthorized.Inc()
				return
			}
		case "GET":
			if user, password, ok := r.BasicAuth(); ok && userR == user && keyR == password {
				var b bytes.Buffer
				res := f(r, w.Header(), &b)
				t.Stop()

				switch res.code {
				case http.StatusOK:
					b.WriteTo(w)

					t.Track()
					if t.Taken() > 500 {
						log.Printf("%s took %d ms to handle %s", id, t.Taken(), r.URL.Path)
					}

					mtrapp.StatusOK.Inc()
				case http.StatusBadRequest:
					http.Error(w, res.msg, res.code)
					mtrapp.StatusBadRequest.Inc()
				case http.StatusInternalServerError:
					http.Error(w, res.msg, res.code)
					mtrapp.StatusInternalServerError.Inc()
					log.Printf("500 serving GET %s %s", r.URL, res.msg)
				default:
					http.Error(w, res.msg, res.code)
				}
			} else {
				w.Header().Set("WWW-Authenticate", "Basic realm=\"GeoNet MTR\"")
				http.Error(w, "Access denied", http.StatusUnauthorized)
				mtrapp.StatusUnauthorized.Inc()
				return
			}

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
