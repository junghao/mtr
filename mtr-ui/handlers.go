package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type page struct {
	// members must be public for reflection
	Body   []byte
	Border border
}

type border struct {
	Title   string
	TagList []string
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
type requestHandler func(r *http.Request, h http.Header, b *bytes.Buffer) *result

var (
	userW, keyW      string
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

func init() {
	userW = os.Getenv("MTR_USER")
	keyW = os.Getenv("MTR_KEY")
}

// For all page structs, get a list of all unique tags and save them in the page struct.
// Enables the use of a simple typeahead search using html5 and datalist.
func (p *page) populateTags() (err error) {
	u := *mtrApiUrl
	u.Path = "/field/tag"
	if p.Border.TagList, err = getAllTagIDs(u.String()); err != nil {
		return err
	}

	return nil
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

func getBytes(urlString string, accept string) (body []byte, err error) {
	var client = &http.Client{}
	var request *http.Request
	var response *http.Response

	if request, err = http.NewRequest("GET", urlString, nil); err != nil {
		return nil, err
	}
	//request.SetBasicAuth(userW, keyW)
	request.Header.Add("Accept", accept)

	if response, err = client.Do(request); err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Wrong response code for %s got %d expected %d", urlString, response.StatusCode, http.StatusOK)
	}

	// Read body, could use io.LimitReader() to avoid a massive read (unlikely)
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// fetch all unique "Tag"s from the mtr-api and return an unordered slice of strings and err
func getAllTagIDs(urlString string) (tagIDs []string, err error) {

	type tagIdBody struct {
		Tag string
	}
	var tagStructs []tagIdBody

	body, err := getBytes(urlString, "application/json;version=1")
	if err != nil {
		return nil, err
	}

	// make a slice of structs {'Tag': string} to populate from json and construct a slice of strings instead
	if err = json.Unmarshal(body, &tagStructs); err != nil {
		return nil, err
	}

	for _, value := range tagStructs {
		tagIDs = append(tagIDs, value.Tag)
	}

	return tagIDs, nil
}

// Very simple toHandler.  Might be able to use the one from mtr-api.
func toHandler(f requestHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var res *result
		var b bytes.Buffer

		// the default content type, wrapped functions can overload if necessary
		w.Header().Set("Content-Type", "text/html")

		switch r.Method {
		case "GET":
			res = f(r, w.Header(), &b)
		case "POST":
			if r.URL.Path == "/search" {
				res = f(r, w.Header(), &b)
			} else {
				http.Error(w, res.msg, res.code)
				log.Printf("improper POST from %s %s", r.URL, res.msg)
				return
			}

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}

		switch res.code {
		case http.StatusOK:
			b.WriteTo(w)
		case http.StatusInternalServerError:
			http.Error(w, res.msg, res.code)
			log.Printf("500 serving GET %s %s", r.URL, res.msg)
		default:
			http.Error(w, res.msg, res.code)
			log.Printf("error serving %s msg: %s", r.URL, res.msg)
		}
	}
}
