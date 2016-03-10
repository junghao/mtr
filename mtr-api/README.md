# MTR Server

A server for metrics.

## Development

Needs a Postgis DB.

## Production

Variables that need setting in the environment for production should be added to `env.list`.

For development add these vars to the env at run time by:

```
export $(cat env.list | grep = | xargs) && go test

or

go build && (export $(cat env.list | grep = | xargs) && ./mtr-api )
```

## Field Metrics

Request should be made over HTTPS.

## API

For API endpoints and methods please refer to `field_metric_test.go`.

