package main

import (
	"encoding/csv"
	"fmt"
	"github.com/GeoNet/mtr/internal"
	wt "github.com/GeoNet/weft/wefttest"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func addData(r wt.Request, t *testing.T) {
	if _, err := r.Do(testServer.URL); err != nil {
		t.Error(err)
	}
}

func compareCsvData(b []byte, expected [][]string, t *testing.T) {
	// for all lines past 0 parse and check values.
	c := csv.NewReader(strings.NewReader(string(b)))
	observed, err := c.ReadAll()
	if err == io.EOF {
		t.Error(err)
	}

	if len(observed) == 0 {
		t.Errorf("CSV file is empty")
	}

	if len(observed) != len(expected) {
		t.Errorf("Number of lines in observed differs from expected %d %d", len(observed), len(expected))
	}

	for i, record := range observed {
		if i == 0 {
			continue
		}

		for f, field := range record {
			if field != expected[i][f] {
				t.Errorf("expected '%s' but observed: '%s' (field %d)",
					strings.Join(expected[i], ", "), strings.Join(observed[i], ", "), f)
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

	now := time.Now().UTC()
	testCounterData := []testPoint{
		testPoint{typeID: http.StatusOK, count: 1.0, time: now},
		testPoint{typeID: http.StatusBadRequest, count: 2.0, time: now}, // add a different typeID at the same time as previous typeID
		testPoint{typeID: http.StatusNotFound, count: 1.0, time: now.Add(time.Second)},
		testPoint{typeID: http.StatusBadRequest, count: 2.0, time: now.Add(time.Second * 2)},
		testPoint{typeID: http.StatusInternalServerError, count: 3.0, time: now.Add(time.Second * 5)},
	}

	// the expected CSV data, ignoring the header fields on the first line
	expectedVals := [][]string{
		[]string{""}, // header line, ignored in test.  Should be time, statusOK, statusBadRequest, StatusNotFound, StatusInternalServerError
		[]string{testCounterData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", testCounterData[0].count), fmt.Sprintf("%.2f", testCounterData[1].count), "", ""},
		[]string{testCounterData[2].time.Format(DYGRAPH_TIME_FORMAT), "", "", fmt.Sprintf("%.2f", testCounterData[2].count), ""},
		[]string{testCounterData[3].time.Format(DYGRAPH_TIME_FORMAT), "", fmt.Sprintf("%.2f", testCounterData[3].count), "", ""},
		[]string{testCounterData[4].time.Format(DYGRAPH_TIME_FORMAT), "", "", "", fmt.Sprintf("%.2f", testCounterData[4].count)},
	}

	for _, td := range testCounterData {
		r.URL = fmt.Sprintf("/application/counter?applicationID=test-app&instanceID=test-instance&typeID=%d&count=%d&time=%s",
			td.typeID, int(td.count), td.time.Format(time.RFC3339))

		addData(r, t)
	}

	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=counters", Method: "GET", Accept: "text/csv"}

	var b []byte
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}
	compareCsvData(b, expectedVals, t)
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

	now := time.Now().UTC()
	timerTestData := []timerTest{
		timerTest{appId: "func-name", count: 1, average: 30, fifty: 73, ninety: 81, time: now},
		timerTest{appId: "func-name2", count: 3, average: 32, fifty: 57, ninety: 59, time: now}, // same time as above but different appId
		timerTest{appId: "func-name3", count: 6, average: 31, fifty: 76, ninety: 82, time: now},
		timerTest{appId: "func-name", count: 4, average: 36, fifty: 73, ninety: 78, time: now.Add(time.Second * 2)},
		timerTest{appId: "func-name", count: 2, average: 33, fifty: 76, ninety: 93, time: now.Add(time.Second * 3)},
		timerTest{appId: "func-name", count: 9, average: 38, fifty: 73, ninety: 91, time: now.Add(time.Second * 7)},
	}

	// the expected CSV data, ignoring the header fields on the first line
	expectedTimerVals := [][]string{
		[]string{""}, // header line, ignored in test.  Should be: time, func-name, func-name2, func-name3.  Only one measurement per metric
		[]string{timerTestData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", timerTestData[0].ninety),
			fmt.Sprintf("%.2f", timerTestData[1].ninety), fmt.Sprintf("%.2f", timerTestData[2].ninety)},
		[]string{timerTestData[3].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", timerTestData[3].ninety), "", ""},
		[]string{timerTestData[4].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", timerTestData[4].ninety), "", ""},
		[]string{timerTestData[5].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", timerTestData[5].ninety), "", ""},
	}

	// Add timer values
	for _, tv := range timerTestData {
		r.URL = fmt.Sprintf("/application/timer?applicationID=test-app&instanceID=test-instance&sourceID=%s&count=%d&average=%d&fifty=%d&ninety=%d&time=%s",
			tv.appId, int(tv.count), int(tv.average), int(tv.fifty), int(tv.ninety), tv.time.Format(time.RFC3339))

		addData(r, t)
	}

	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=timers", Method: "GET", Accept: "text/csv"}

	var b []byte
	var err error
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}
	compareCsvData(b, expectedTimerVals, t)

	// do same test with sourceID specified since it uses another SQL query and outputs different results
	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=timers&sourceID=func-name", Method: "GET", Accept: "text/csv"}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	expectedTimerSrcVals := [][]string{
		[]string{""}, // header line, ignored in test.  Should be: time, func-name.
		[]string{timerTestData[0].time.Format(DYGRAPH_TIME_FORMAT),
			fmt.Sprintf("%.2f", timerTestData[0].average),
			fmt.Sprintf("%.2f", timerTestData[0].fifty),
			fmt.Sprintf("%.2f", timerTestData[0].ninety)},
		[]string{timerTestData[3].time.Format(DYGRAPH_TIME_FORMAT),
			fmt.Sprintf("%.2f", timerTestData[3].average),
			fmt.Sprintf("%.2f", timerTestData[3].fifty),
			fmt.Sprintf("%.2f", timerTestData[3].ninety)},
		[]string{timerTestData[4].time.Format(DYGRAPH_TIME_FORMAT),
			fmt.Sprintf("%.2f", timerTestData[4].average),
			fmt.Sprintf("%.2f", timerTestData[4].fifty),
			fmt.Sprintf("%.2f", timerTestData[4].ninety)},
		[]string{timerTestData[5].time.Format(DYGRAPH_TIME_FORMAT),
			fmt.Sprintf("%.2f", timerTestData[5].average),
			fmt.Sprintf("%.2f", timerTestData[5].fifty),
			fmt.Sprintf("%.2f", timerTestData[5].ninety)},
	}
	compareCsvData(b, expectedTimerSrcVals, t)

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

	//"/application/metric?applicationID=test-app&instanceID=test-instance&typeID=1000&value=10000&time=2015-05-14T21:40:30Z"
	//applicationID=test-app   instanceID=test-instance    typeID=1000    value=10000    time=2015-05-14T21:40:30Z"
	now := time.Now().UTC()
	memTestData := []memoryTest{
		memoryTest{appId: "test-app", instanceId: "test-instance", typeId: 1000, value: 10, time: now},
		memoryTest{appId: "test-app", instanceId: "test-instance", typeId: 1000, value: 9, time: now.Add(time.Second)},
		memoryTest{appId: "test-app", instanceId: "test-instance", typeId: 1000, value: 8, time: now.Add(time.Second * 2)},
		memoryTest{appId: "test-app", instanceId: "test-instance", typeId: 1000, value: 7, time: now.Add(time.Second * 3)},
		memoryTest{appId: "test-app", instanceId: "test-instance", typeId: 1000, value: 6, time: now.Add(time.Second * 6)},
	}

	// the expected CSV data, ignoring the header fields on the first line
	expectedMemVals := [][]string{
		[]string{""}, // header line, ignored in test.
		[]string{memTestData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", memTestData[0].value)},
		[]string{memTestData[1].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", memTestData[1].value)},
		[]string{memTestData[2].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", memTestData[2].value)},
		[]string{memTestData[3].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", memTestData[3].value)},
		[]string{memTestData[4].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", memTestData[4].value)},
	}

	// Add timer values
	for _, mt := range memTestData {
		// /application/metric?applicationID=test-app&instanceID=test-instance&typeID=1000&value=10000&time=2015-05-14T21:40:30Z
		r.URL = fmt.Sprintf("/application/metric?applicationID=%s&instanceID=%s&typeID=%d&value=%d&time=%s",
			mt.appId, mt.instanceId, mt.typeId, int(mt.value), mt.time.Format(time.RFC3339))

		addData(r, t)
	}

	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=memory", Method: "GET", Accept: "text/csv"}

	var err error
	var b []byte
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	compareCsvData(b, expectedMemVals, t)
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

	type memoryTest struct {
		appId, instanceId string
		typeId            int
		value             float64
		time              time.Time
	}

	// handling objects and routines in the same test since it's the same method being exercised
	now := time.Now().UTC()
	objTestData := []memoryTest{
		memoryTest{appId: "test-app", instanceId: "test-instance", typeId: int(internal.MemHeapObjects), value: 8, time: now.Add(time.Second)},
		memoryTest{appId: "test-app", instanceId: "test-instance", typeId: int(internal.MemHeapObjects), value: 12, time: now.Add(time.Second * 2)},
		memoryTest{appId: "test-app", instanceId: "test-instance", typeId: int(internal.Routines), value: 1, time: now.Add(time.Second * 3)},
		memoryTest{appId: "test-app", instanceId: "test-instance", typeId: int(internal.Routines), value: 3, time: now.Add(time.Second * 4)},
		memoryTest{appId: "test-app", instanceId: "test-instance", typeId: int(internal.MemSys), value: 10, time: now.Add(time.Second * 5)},
		memoryTest{appId: "test-app", instanceId: "test-instance", typeId: int(internal.MemHeapAlloc), value: 9, time: now.Add(time.Second * 6)},
		memoryTest{appId: "test-app", instanceId: "test-instance", typeId: int(internal.MemHeapSys), value: 7, time: now.Add(time.Second * 7)},
	}

	// the expected CSV data, ignoring the header fields on the first line
	expectedObjValues := [][]string{
		[]string{""}, // header line, ignored in test.
		[]string{objTestData[0].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", objTestData[0].value)},
		[]string{objTestData[1].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", objTestData[1].value)},
	}

	// Add timer values
	for _, ov := range objTestData {
		// /application/metric?applicationID=test-app&instanceID=test-instance&typeID=1000&value=10000&time=2015-05-14T21:40:30Z
		r.URL = fmt.Sprintf("/application/metric?applicationID=%s&instanceID=%s&typeID=%d&value=%d&time=%s",
			ov.appId, ov.instanceId, ov.typeId, int(ov.value), ov.time.Format(time.RFC3339))

		addData(r, t)
	}

	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=objects", Method: "GET", Accept: "text/csv"}

	var err error
	var b []byte
	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	compareCsvData(b, expectedObjValues, t)

	// test again for number of goroutines

	expectedRoutineValues := [][]string{
		[]string{""}, // header line, ignored in test.
		[]string{objTestData[2].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", objTestData[2].value)},
		[]string{objTestData[3].time.Format(DYGRAPH_TIME_FORMAT), fmt.Sprintf("%.2f", objTestData[3].value)},
	}

	r = wt.Request{ID: wt.L(), URL: "/app/metric?applicationID=test-app&group=routines", Method: "GET", Accept: "text/csv"}

	if b, err = r.Do(testServer.URL); err != nil {
		t.Error(err)
	}

	compareCsvData(b, expectedRoutineValues, t)
}
