# MTR API

The API server for collecting metrics.

## Development

* Needs a Postgis DB.
* There are tests to add a small amount of test data and check routes.

## Field Metrics

Request should be made over HTTPS.

## API Docs

For API endpoints and methods please refer to `field_metric_test.go`.

# MTR UI

The web based user interface for MTR.

# Protobufs

* `protobuf` contains protobuf definitions for mtr-api and clients.
* Update the Go pkg in `mtrpb` by compiling the protobuf using the Go plugin (https://github.com/golang/protobuf) and:

```
protoc --proto_path=protobuf/mtrpb/ --go_out=mtrpb protobuf/mtrpb/*
```

* If you run into errors check the version of protobuf added by govendor, groupcache may add an older version.