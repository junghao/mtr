package main

import (
	"fmt"
	"github.com/GeoNet/mtr/mtrpb"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"net/http"
	"os"
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

var userW, keyW string

func init() {
	userW = os.Getenv("MTR_USER")
	keyW = os.Getenv("MTR_KEY")
}

// For all page structs, get a list of all unique tags and save them in the page struct.
// Enables the use of a simple typeahead search using html5 and datalist.
func (p *page) populateTags() (err error) {
	u := *mtrApiUrl
	u.Path = "/tag"
	if p.Border.TagList, err = getAllTagIDs(u.String()); err != nil {
		return err
	}

	return nil
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
		msg := ""
		if response.Body != nil {
			if b, err := ioutil.ReadAll(response.Body); err == nil {
				msg = ":" + string(b)
			}
		}

		return nil, fmt.Errorf("Wrong response code for %s got %d expected %d %s", urlString, response.StatusCode, http.StatusOK, msg)
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
	b, err := getBytes(urlString, "application/x-protobuf")
	if err != nil {
		return nil, err
	}

	var tr mtrpb.TagResult

	if err = proto.Unmarshal(b, &tr); err != nil {
		return nil, err
	}

	if tr.Result != nil {
		for _, value := range tr.Result {
			tagIDs = append(tagIDs, value.Tag)
		}
	}

	return tagIDs, nil
}
