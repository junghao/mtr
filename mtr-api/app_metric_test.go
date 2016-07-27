package main

import (
	"encoding/csv"
	"fmt"
	"github.com/GeoNet/mtr/internal"
	wt "github.com/GeoNet/weft/wefttest"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func addData(r wt.Request, t *testing.T) {
	if _, err := r.Do(testServer.URL); err != nil {
		t.Error(err)
	}
}

// loc returns a string representing the line of code 2 functions calls back.
func loc() (loc string) {
	_, _, l, _ := runtime.Caller(2)
	return "L" + strconv.Itoa(l)
}

func compareCsvData(b []byte, expected [][]string, t *testing.T) {
	l := loc()
	// for all lines past 0 parse and check values.
	c := csv.NewReader(strings.NewReader(string(b)))
	observed, err := c.ReadAll()
	if err == io.EOF {
		t.Error(err)
	}

	if len(observed) == 0 {
		t.Errorf("%s CSV file is empty", l)
	}

	if len(observed) != len(expected) {
		t.Errorf("%s Number of lines in observed: %d differs from expected: %d", l, len(observed), len(expected))
	}

	for i, record := range observed {
		if i == 0 {
			continue
		}

		if len(record) != len(expected[i]) {
			t.Errorf("%s length of record %d not equal to expected %d", l, len(record), len(expected))
		}

		for f, field := range record {
			if field != expected[i][f] {
				t.Errorf("%s expected '%s' but observed: '%s' (field %d)",
					l, strings.Join(expected[i], ", "), strings.Join(observed[i], ", "), f)
			}
		}
	}
}

func TestAppMetricCounterCsv(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	r := wt.Request{
		User:     userW,
		Password: keyW,
		Method:   "PUT",
	}
	var err error

	type testPoint struct {
		typeID int
		count  float64
		time   time.Time
	}

	// Testing the "counter" group

	utcNow := time.Now().UTC().Truncate(time.Second)
	t0 := utcNow.Add(time.Second * -10)
	testCounterData := []testPoint{
		{typeID: http.StatusOK, count: 1.0, time: t0},
		{typeID: http.StatusBadRequest, count: 2.0, time: t0}, // add a different typeID at the same time as previous typeID
		{typeID: http.StatusNotFound, count: 1.0, time: t0.Add(time.Second * 2)},
		{typeID: http.StatusBadRequest, count: 2.0, time: t0.Add(time.Second * 4)},
		{typeID: http.StatusInternalServerError, count: 3.0, time: t0.Add(time.Second * 6)},
	}

	// the expected CSV data, ignoring the header fields on the first line
	expectedVals := [][]string{
		{""}, // header line, ignored in test.  Should be time, statusOK, statusBadRequest, StatusNotFound, StatusInternalServerError
		{testCounterData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", testCounterData[0].count), fmt.Sprintf("%.2f", testCounterData[1].count), "", ""},
		{testCounterData[2].time.Format(DYGRAPH_TIME_FORMAT), "", "", fmt.Sprintf("%.2f", testCounterData[2].count), ""},
		{testCounterData[3].time.Format(DYGRAPH_TIME_FORMAT), "", fmt.Sprintf("%.2f", testCounterData[3].count), "", ""},
		{testCounterData[4].time.Format(DYGRAPH_TIME_FORMAT), "", "", "", fmt.Sprintf("%.2f", testCounterData[4].count)},
	}

	for _, td := range testCounterData {
		r.URL = fmt.Sprintf("/application/counter?applicationID=test-app&instanceID=test-instance&typeID=%d&count=%d&time=%s",
			td.typeID, int(td.count), td.time.Format(time.RFC3339))

		addData(r, t)
	}

	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=counters&resolution=full", Method: "GET", Accept: "text/csv"}

	var b []byte
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}
	compareCsvData(b, expectedVals, t)

	// test with time range specified
	expectedSubset := [][]string{
		{""}, // header line, ignored in test.  Should be time, statusOK, statusBadRequest
		{testCounterData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", testCounterData[0].count), fmt.Sprintf("%.2f", testCounterData[1].count)},
	}

	// time window so we only get the points at t0.  RFC3339 has second precision.
	start := t0.Add(time.Second * -1).Format(time.RFC3339)
	end := t0.Add(time.Second).Format(time.RFC3339)
	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=counters&resolution=full&startDate=" + start + "&endDate=" + end, Method: "GET", Accept: "text/csv"}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}
	compareCsvData(b, expectedSubset, t)
}

func TestAppMetricTimerCsv(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	// Testing the "timers" group, could move to another testing function
	r := wt.Request{
		User:     userW,
		Password: keyW,
		Method:   "PUT",
	}

	type timerTest struct {
		appId   string
		count   float64
		average float64
		fifty   float64
		ninety  float64
		time    time.Time
	}

	utcNow := time.Now().UTC().Truncate(time.Second)
	t0 := utcNow.Add(time.Second * -10)
	timerTestData := []timerTest{
		{appId: "func-name", count: 1, average: 30, fifty: 73, ninety: 81, time: t0},
		{appId: "func-name2", count: 3, average: 32, fifty: 57, ninety: 59, time: t0}, // same time as above but different appId
		{appId: "func-name3", count: 6, average: 31, fifty: 76, ninety: 82, time: t0},
		{appId: "func-name", count: 4, average: 36, fifty: 73, ninety: 78, time: t0.Add(time.Second * 2)},
		{appId: "func-name", count: 2, average: 33, fifty: 76, ninety: 93, time: t0.Add(time.Second * 4)},
		{appId: "func-name", count: 9, average: 38, fifty: 73, ninety: 91, time: t0.Add(time.Second * 6)},
	}

	// the expected CSV data, ignoring the header fields on the first line
	expectedTimerVals := [][]string{
		{""}, // header line, ignored in test.  Should be: time, func-name, func-name2, func-name3.  Only one measurement per metric
		{timerTestData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", timerTestData[0].ninety),
			fmt.Sprintf("%.2f", timerTestData[1].ninety), fmt.Sprintf("%.2f", timerTestData[2].ninety)},
		{timerTestData[3].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", timerTestData[3].ninety), "", ""},
		{timerTestData[4].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", timerTestData[4].ninety), "", ""},
		{timerTestData[5].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", timerTestData[5].ninety), "", ""},
	}

	// Add timer values
	for _, tv := range timerTestData {
		r.URL = fmt.Sprintf("/application/timer?applicationID=test-app&instanceID=test-instance&sourceID=%s&count=%d&average=%d&fifty=%d&ninety=%d&time=%s",
			tv.appId, int(tv.count), int(tv.average), int(tv.fifty), int(tv.ninety), tv.time.Format(time.RFC3339))

		addData(r, t)
	}

	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=timers&resolution=full", Method: "GET", Accept: "text/csv"}

	var b []byte
	var err error
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}
	compareCsvData(b, expectedTimerVals, t)

	// same test with time range
	start := timerTestData[3].time.Add(time.Second * -1).Format(time.RFC3339)
	end := timerTestData[3].time.Add(time.Second).Format(time.RFC3339)
	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=timers&resolution=full&startDate=" + start + "&endDate=" + end, Method: "GET", Accept: "text/csv"}

	expectedTimerSubset := [][]string{
		{""}, // header line, ignored in test.  Should be: time, func-name.
		{timerTestData[3].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", timerTestData[3].ninety)},
	}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	//fmt.Println(string(b), start, end, "subset time", timerTestData[3].time)
	compareCsvData(b, expectedTimerSubset, t)

	// do same test with sourceID specified since it uses another SQL query and outputs different results (average, fifty, ninety for the specified applicationID)
	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=timers&sourceID=func-name&resolution=full", Method: "GET", Accept: "text/csv"}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	expectedTimerSrcVals := [][]string{
		{""}, // header line, ignored in test.  Should be: time, func-name.
		{timerTestData[0].time.Format(DYGRAPH_TIME_FORMAT),
			fmt.Sprintf("%.2f", timerTestData[0].average),
			fmt.Sprintf("%.2f", timerTestData[0].fifty),
			fmt.Sprintf("%.2f", timerTestData[0].ninety)},
		{timerTestData[3].time.Format(DYGRAPH_TIME_FORMAT),
			fmt.Sprintf("%.2f", timerTestData[3].average),
			fmt.Sprintf("%.2f", timerTestData[3].fifty),
			fmt.Sprintf("%.2f", timerTestData[3].ninety)},
		{timerTestData[4].time.Format(DYGRAPH_TIME_FORMAT),
			fmt.Sprintf("%.2f", timerTestData[4].average),
			fmt.Sprintf("%.2f", timerTestData[4].fifty),
			fmt.Sprintf("%.2f", timerTestData[4].ninety)},
		{timerTestData[5].time.Format(DYGRAPH_TIME_FORMAT),
			fmt.Sprintf("%.2f", timerTestData[5].average),
			fmt.Sprintf("%.2f", timerTestData[5].fifty),
			fmt.Sprintf("%.2f", timerTestData[5].ninety)},
	}
	compareCsvData(b, expectedTimerSrcVals, t)

	// similar test but with a time range
	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=timers&sourceID=func-name&resolution=full&startDate=" + start + "&endDate=" + end, Method: "GET", Accept: "text/csv"}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	expectedTimerSrcSubset := [][]string{
		{""}, // header line, ignored in test.  Should be: time, func-name.
		{timerTestData[3].time.Format(DYGRAPH_TIME_FORMAT),
			fmt.Sprintf("%.2f", timerTestData[3].average),
			fmt.Sprintf("%.2f", timerTestData[3].fifty),
			fmt.Sprintf("%.2f", timerTestData[3].ninety)},
	}
	compareCsvData(b, expectedTimerSrcSubset, t)

}

func TestAppMetricMemoryCsv(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	// Testing the "timers" group, could move to another testing function
	r := wt.Request{
		User:     userW,
		Password: keyW,
		Method:   "PUT",
	}

	type memoryTest struct {
		appId, instanceId string
		typeId            int
		value             float64
		time              time.Time
	}

	utcNow := time.Now().UTC().Truncate(time.Second)
	t0 := utcNow.Add(time.Second * -10)
	memTestData := []memoryTest{
		{appId: "test-app", instanceId: "test-instance", typeId: 1000, value: 10, time: t0},
		{appId: "test-app", instanceId: "test-instance", typeId: 1000, value: 9, time: t0.Add(time.Second * 2)},
		{appId: "test-app", instanceId: "test-instance", typeId: 1000, value: 8, time: t0.Add(time.Second * 4)},
		{appId: "test-app", instanceId: "test-instance", typeId: 1000, value: 7, time: t0.Add(time.Second * 6)},
		{appId: "test-app", instanceId: "test-instance", typeId: 1000, value: 6, time: t0.Add(time.Second * 8)},
	}

	// the expected CSV data, ignoring the header fields on the first line
	expectedMemVals := [][]string{
		{""}, // header line, ignored in test.
		{memTestData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", memTestData[0].value)},
		{memTestData[1].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", memTestData[1].value)},
		{memTestData[2].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", memTestData[2].value)},
		{memTestData[3].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", memTestData[3].value)},
		{memTestData[4].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", memTestData[4].value)},
	}

	// Add timer values
	for _, mt := range memTestData {
		// /application/metric?applicationID=test-app&instanceID=test-instance&typeID=1000&value=10000&time=2015-05-14T21:40:30Z
		r.URL = fmt.Sprintf("/application/metric?applicationID=%s&instanceID=%s&typeID=%d&value=%d&time=%s",
			mt.appId, mt.instanceId, mt.typeId, int(mt.value), mt.time.Format(time.RFC3339))

		addData(r, t)
	}

	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=memory&resolution=full", Method: "GET", Accept: "text/csv"}

	var err error
	var b []byte
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	compareCsvData(b, expectedMemVals, t)

	// test with time range
	start := memTestData[3].time.Add(time.Second * -1).UTC().Format(time.RFC3339)
	end := memTestData[3].time.Add(time.Second).UTC().Format(time.RFC3339)
	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=memory&resolution=full&startDate=" + start + "&endDate=" + end, Method: "GET", Accept: "text/csv"}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	expectedMemSubset := [][]string{
		{""}, // header line, ignored in test.
		{memTestData[3].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", memTestData[3].value)},
	}
	compareCsvData(b, expectedMemSubset, t)

}

func TestAppMetricObjectsCsv(t *testing.T) {
	setup(t)
	defer teardown()

	// Load test data.
	if err := routes.DoAllStatusOk(testServer.URL); err != nil {
		t.Error(err)
	}

	// Testing the "timers" group, could move to another testing function
	r := wt.Request{
		User:     userW,
		Password: keyW,
		Method:   "PUT",
	}

	type objectTest struct {
		appId, instanceId string
		typeId            int
		value             float64
		time              time.Time
	}

	// handling objects and routines in the same test since it's the same method being exercised
	utcNow := time.Now().UTC().Truncate(time.Second)
	t0 := utcNow.Add(time.Second * -10)
	objTestData := []objectTest{
		{appId: "test-app", instanceId: "test-instance", typeId: int(internal.MemHeapObjects), value: 8, time: t0.Add(time.Second)},
		{appId: "test-app", instanceId: "test-instance", typeId: int(internal.MemHeapObjects), value: 12, time: t0.Add(time.Second * 2)},
		{appId: "test-app", instanceId: "test-instance", typeId: int(internal.Routines), value: 1, time: t0.Add(time.Second * 3)},
		{appId: "test-app", instanceId: "test-instance", typeId: int(internal.Routines), value: 3, time: t0.Add(time.Second * 6)},
		{appId: "test-app", instanceId: "test-instance", typeId: int(internal.MemSys), value: 10, time: t0.Add(time.Second * 7)},
		{appId: "test-app", instanceId: "test-instance", typeId: int(internal.MemHeapAlloc), value: 9, time: t0.Add(time.Second * 8)},
		{appId: "test-app", instanceId: "test-instance", typeId: int(internal.MemHeapSys), value: 7, time: t0.Add(time.Second * 9)},
	}

	// the expected CSV data, ignoring the header fields on the first line
	expectedObjValues := [][]string{
		{""}, // header line, ignored in test.
		{objTestData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", objTestData[0].value)},
		{objTestData[1].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", objTestData[1].value)},
	}

	// Add timer values
	for _, ov := range objTestData {
		// /application/metric?applicationID=test-app&instanceID=test-instance&typeID=1000&value=10000&time=2015-05-14T21:40:30Z
		r.URL = fmt.Sprintf("/application/metric?applicationID=%s&instanceID=%s&typeID=%d&value=%d&time=%s",
			ov.appId, ov.instanceId, ov.typeId, int(ov.value), ov.time.Format(time.RFC3339))

		addData(r, t)
	}

	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=objects&resolution=full", Method: "GET", Accept: "text/csv"}

	var err error
	var b []byte
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	compareCsvData(b, expectedObjValues, t)

	// test again for number of goroutines

	expectedRoutineValues := [][]string{
		{""}, // header line, ignored in test.
		{objTestData[2].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", objTestData[2].value)},
		{objTestData[3].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", objTestData[3].value)},
	}

	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=routines&resolution=full", Method: "GET", Accept: "text/csv"}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	compareCsvData(b, expectedRoutineValues, t)

	// Also test for invalid applicationID, should give a 404
	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=NOT_AN_APP&group=routines&resolution=full", Method: "GET", Accept: "text/csv", Status: http.StatusNotFound}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	// Test again with time range
	start := objTestData[3].time.Add(time.Second * -1).UTC().Format(time.RFC3339)
	end := objTestData[3].time.Add(time.Second).UTC().Format(time.RFC3339)
	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=routines&resolution=full&startDate=" + start + "&endDate=" + end, Method: "GET", Accept: "text/csv"}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	expectedRoutineSubset := [][]string{
		{""}, // header line, ignored in test.
		{objTestData[3].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", objTestData[3].value)},
	}

	compareCsvData(b, expectedRoutineSubset, t)
}
