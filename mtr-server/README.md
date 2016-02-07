# MTR Server

A server for metrics.

## Development

Needs a Postgis DB.

## Production

Variables that need setting in the environment for production should be added to `env.list`.

## Field Metrics

Request should be made over HTTPS.

### Metric types

Valid `typeID` values are 

* `voltage` - voltage in mV (int32).
* `clock` - clock quality (int32).
* `satellites` - number of satellites tracked (int32).

## REST API

For API endpoints and methods please refer to `field_metric_test.go`.

