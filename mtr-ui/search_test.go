package main

import "testing"

func TestGetMatchingMetrics(t *testing.T) {
	jsonTestOutput := []byte(`[{"TypeID":"voltage", "DeviceID":"companyA", "Tag":"1234"}, {"TypeID":"voltage", "DeviceID":"companyB", "Tag":"ABCD"}]`)
	testServer := jsonTestServer(jsonTestOutput)
	defer jsonTestServerTearDown(testServer)

	matches, err := getMatchingMetrics(testServer.URL)
	if err != nil {
		t.Error(err)
	}

	expectedMetrics := matchingMetrics{metricInfo{TypeID: "voltage", DeviceID: "companyA", Tag: "1234"},
		metricInfo{TypeID: "voltage", DeviceID: "companyB", Tag: "ABCD"}}

	if len(matches) != len(expectedMetrics) {
		t.Errorf("observed metrics length: %d did not match expected length: %d\n", len(matches), len(expectedMetrics))
	}

	for idx, val := range matches {
		expect := expectedMetrics[idx]
		// compare all struct members apart from []bytes
		if val.DeviceID != expect.DeviceID || val.Tag != expect.Tag || val.TypeID != expect.TypeID {
			t.Errorf("observed metric did not match expected for index %d\n", idx)
		}
	}
}
