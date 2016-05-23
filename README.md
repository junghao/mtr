# MTR

Field, data, and application metrics.

## database

Defines the database schema for mtr.  Uses a Postgres+Postgis DB.  Set up a Docker instance for
development:

```
docker run --name hazdb -p 5432:5432 -d  quay.io/geonet/haz:database
./database/scripts/initdb.sh
```


## internal

constants for communication between `mtrapp` and `mtr-api`.


## mtr-api

The HTTP server for collecting metrics.  Needs the `database` set up for development.

### Tests

The API is tested through the web server using an HTTP client see `routes_test.go`.  This should
fully exercise the code (although a test coverage tool may not show this).  It requires
the database to be running.  This also adds test data and tests the more complicated GET
responses (protobuf).  Testing the protobuf responses here means they don't need
testing again in `mtr-ui` and it is easier to manage test data.

`route_test.go` also provides useful API documentation.

Run tests by exporting the env var defined in `env.list` and then:

```
go test
```

For working with time series SVG plots run just one test:

```
go test -run TestPlotData
```

Then compile and run the app:

```
go build && ./mtr-api
```

and visit:

* http://localhost:8080/field/metric?deviceID=gps-taupoairport&typeID=voltage
* http://localhost:8080/data/latency?siteID=TAUP&typeID=latency.strong

### Adding Features

* Prefer URL query parameters over body content for PUT methods for API consistency.  Follow the query parameter naming scheme.
* GET methods should return SVG, Protobuf, or GeoJSON (for use in web maps).

Adding code:

* add database definitions to `database`.
* domain objects to match the database are defined in `domain.go`.  Make additions there.  Read the assumptions documented there as well.
* add methods to domain objects for put, get, delete as required e.g., `data_site.go`.  Keep in mind that non get methods are passed a nil buffer pointer.
* implement weft.RequestHandler in handlers to convert your domain methods to handlers.
* add routes to the mux in `server.go`.
* test the routes in `routes_test.go`.
* if you add services to return protobuf also test the response body in `routes_test.go`
* if you add services to generate SVG plots add a method to generate some test data e.g., `TestPlotData` in `routes_test.go`.

## mtr-ui

Provides a web interface to mtr-api.


## mtrapp

A package for gathering application metrics from Go programs.


## mtrpb

Go protobuf code generated from the definitions in `protobuf`.
Update the Go pkg in `mtrpb` by compiling the protobuf using the Go plugin (https://github.com/golang/protobuf) and:

```
protoc --proto_path=protobuf/mtrpb/ --go_out=mtrpb protobuf/mtrpb/*
```

* If you run into errors check the version of protobuf added by govendor, groupcache may add an older version.


## protobuf

Protobuf definitions for `mtr-api` services.


## ts

Time series SVG plots.