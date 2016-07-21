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
	MapList []mapDef
}

type mapDef struct {
	TypeIDs []string
	ApiUrl  string
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

// get types for map, incl. fieldType, dataType, dataCompletenessType...
func (p *mapPage) populateTypes() (err error) {
	u := *mtrApiUrl
	//1. field types
	u.Path = "/field/type"
	fieldMap := mapDef{ApiUrl: p.MtrApiUrl + "/field/metric/summary?bbox=NewZealand&width=800"}

	if fieldMap.TypeIDs, err = getAllFieldTypes(u.String()); err != nil {
		return err
	}
	p.Border.MapList = append(p.Border.MapList, fieldMap)

	//2. data types
	u.Path = "/data/type"
	dataMap := mapDef{ApiUrl: p.MtrApiUrl + "/data/latency/summary?bbox=NewZealand&width=800"}

	if dataMap.TypeIDs, err = getAllDataTypes(u.String()); err != nil {
		return err
	}
	p.Border.MapList = append(p.Border.MapList, dataMap)

	//3. dataCompleteness types
	u.Path = "/data/completeness/type"
	dataCompletenessMap := mapDef{ApiUrl: p.MtrApiUrl + "/data/completeness/summary?bbox=NewZealand&width=800"}
	//use same function for dataType as they all return mtrpb.DataTypeResult
	if dataCompletenessMap.TypeIDs, err = getAllDataTypes(u.String()); err != nil {
		return err
	}
	p.Border.MapList = append(p.Border.MapList, dataCompletenessMap)

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

// fetch all field "typeIDs"s from the mtr-api and return an unordered slice of strings and err
func getAllFieldTypes(urlString string) (typeIDs []string, err error) {
	b, err := getBytes(urlString, "application/x-protobuf")
	if err != nil {
		return nil, err
	}

	var ftr mtrpb.FieldTypeResult

	if err = proto.Unmarshal(b, &ftr); err != nil {
		return nil, err
	}

	if ftr.Result != nil {
		for _, value := range ftr.Result {
			typeIDs = append(typeIDs, value.TypeID)
		}
	}

	return typeIDs, nil
}

// fetch all data "typeIDs"s from the mtr-api and return an unordered slice of strings and err
func getAllDataTypes(urlString string) (typeIDs []string, err error) {
	b, err := getBytes(urlString, "application/x-protobuf")
	if err != nil {
		return nil, err
	}

	var dtr mtrpb.DataTypeResult

	if err = proto.Unmarshal(b, &dtr); err != nil {
		return nil, err
	}

	if dtr.Result != nil {
		for _, value := range dtr.Result {
			typeIDs = append(typeIDs, value.TypeID)
		}
	}

	return typeIDs, nil
}
