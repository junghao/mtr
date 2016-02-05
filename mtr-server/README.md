# MTR Server

A server for metrics.

## Development

Needs a Postgis DB.

Compile, set env, and run the server:

```
go build && (export $(cat env.list | grep = | xargs) && ./mtr-server )
```

## Field Metrics

Request should be made over HTTPS.

### Metric types

Valid `typeID` values are 

* `voltage` - voltage in mV (int32).
* `clock` - clock quality (int32).
* `satellites` - number of satellites tracked (int32).


### Locality 

Creates a locality.

```
curl -u test:test -X PUT "http://localhost:8080/field/locality?localityID=taupoairport&name=Taupo+Airport&latitude=-38.74270&longitude=176.08100"
```

Deletes the locality and any associated metrics.

```
curl -u test:test -X DELETE "http://localhost:8080/field/locality?localityID=taupoairport"
```

### Model

Creates an equipment model.

```
curl -u test:test -X PUT "http://localhost:8080/field/model?modelID=Trimble+NetR9"
```

Deletes an equipment model and any metrics for it

```
curl -u test:test -X DELETE "http://localhost:8080/field/model?modelID=Trimble+NetR9"
```

### Site

Creates a site with a code. 

```
curl -u test:test -X PUT "http://localhost:8080/field/site?code=TAUP"
```

Deletes a site and removes the code from any associated metrics.  Does not delete the metrics.

```
curl -u test:test -X DELETE "http://localhost:8080/field/site?code=TAUP"
```

### Metrics

Creates a metric.  `code` is optional.  It is an error to send a metric again.  It is an error to send more than one metric per hour.

```
curl -u test:test -X PUT "http://localhost:8080/field/metric?localityID=taupoairport&code=TAUP&modelID=Trimble+NetR9&typeID=voltage&time=2016-02-03T08:00:16Z&value=14100"
```

Delete all metrics for a model at a locality

```
curl -u test:test -X DELETE "http://localhost:8080/field/metric?localityID=taupoairport&code=TAUP&modelID=Trimble+NetR9"
```
