package main

import (
	"net/http"
	"bytes"
	"log"
	"strings"
)

type page struct {
	// members must be public for reflection
	Title string
	Body  []byte
}

type tagsPage struct {
	page
	Tags  tags
}

type tags []tag

type tag struct {
	TypeID   string
	DeviceID string
	Tag      string
}

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
type requestHandler func(r *http.Request, w http.ResponseWriter, b *bytes.Buffer) *result

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

func notFoundError(message string) *result {
	return &result{ok: false, code: http.StatusNotFound, msg: message}
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

func tagHandler(r *http.Request, w http.ResponseWriter, b *bytes.Buffer) *result {

	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	var p tagsPage //{Title:"Tags"}
	p.page.Title = "Tags"

	//if res := p.load(r.URL.Path); !res.ok {
	//	return res

	// TODO: handle requests for a specific tag (in a different func), get the tag from the mtr-api and return a slice of structs
	p.Tags = append(p.Tags, tag{DeviceID:"hello", TypeID:"Something", Tag:"madeup"})
	p.Tags = append(p.Tags, tag{DeviceID:"Yep", TypeID:"More", Tag:"stuff"})

	if err := tagsTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

// example handler.
func handler(r *http.Request, w http.ResponseWriter, b *bytes.Buffer) *result {

	if res := checkQuery(r, []string{}, []string{}); !res.ok {
		return res
	}

	// We create a page struct with variables to substitute into the loaded template
	p := page{Title:"a title"}

	//if res := p.load(r.URL.Path); !res.ok {
	//	return res
	//}
	//
	//if err := demoPageTemplate.ExecuteTemplate(b, "border", p); err != nil {
	//	return internalServerError(err)
	//}


	if err := borderTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return internalServerError(err)
	}

	//b.WriteString("Hello from a demo page")

	return &statusOK
}

// Very simple toHandler.  Might even be able to use the one from mtr-api.
func toHandler(f requestHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// the default content type, wrapped functions can overload if necessary
		w.Header().Set("Content-Type", "text/html")

		switch r.Method {
		case "GET":
			var b bytes.Buffer
			res := f(r, w, &b)

			switch res.code {
			case http.StatusOK:
				b.WriteTo(w)
			case http.StatusInternalServerError:
				http.Error(w, res.msg, res.code)
				log.Printf("500 serving GET %s %s", r.URL, res.msg)
			default:
				http.Error(w, res.msg, res.code)
			}

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

