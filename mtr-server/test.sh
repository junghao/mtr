#!/bin/sh

# These queries can all be repeated and will update the db for non *ID or code parameters.
curl -u test:test -X PUT "http://localhost:8080/field/locality?localityID=taupoairport&name=Taupo+Airport&latitude=-38.74270&longitude=176.08100"
curl -u test:test -X PUT "http://localhost:8080/field/locality?localityID=ahititi&name=Ahititi&latitude=-38.41149&longitude=178.04593"
curl -u test:test -X PUT "http://localhost:8080/field/model?modelID=Trimble+NetR9"
curl -u test:test -X PUT "http://localhost:8080/field/model?modelID=MikroTik+RouterOS"
curl -u test:test -X PUT "http://localhost:8080/field/site?code=TAUP"

# Metrics cannot be repeated.  code is optional.
# voltage is mV
curl -u test:test -X PUT "http://localhost:8080/field/metric?localityID=taupoairport&code=TAUP&modelID=Trimble+NetR9&typeID=voltage&time=2016-02-03T08:00:16Z&value=14100"
curl -u test:test -X PUT "http://localhost:8080/field/metric?localityID=taupoairport&code=TAUP&modelID=Trimble+NetR9&typeID=voltage&time=2016-02-03T09:15:16Z&value=14200"
curl -u test:test -X PUT "http://localhost:8080/field/metric?localityID=taupoairport&code=TAUP&modelID=Trimble+NetR9&typeID=voltage&time=2016-02-03T10:30:16Z&value=14400"
curl -u test:test -X PUT "http://localhost:8080/field/metric?localityID=taupoairport&code=TAUP&modelID=Trimble+NetR9&typeID=voltage&time=2016-02-03T11:45:16Z&value=14900"

# locality with no code.
curl -u test:test -X PUT "http://localhost:8080/field/metric?localityID=ahititi&modelID=MikroTik+RouterOS&typeID=voltage&time=2016-02-03T08:00:16Z&value=12100"
curl -u test:test -X PUT "http://localhost:8080/field/metric?localityID=ahititi&modelID=MikroTik+RouterOS&typeID=voltage&time=2016-02-03T09:15:16Z&value=12200"
curl -u test:test -X PUT "http://localhost:8080/field/metric?localityID=ahititi&modelID=MikroTik+RouterOS&typeID=voltage&time=2016-02-03T10:30:16Z&value=12400"
curl -u test:test -X PUT "http://localhost:8080/field/metric?localityID=ahititi&modelID=MikroTik+RouterOS&typeID=voltage&time=2016-02-03T11:45:16Z&value=12900"

# Deleteing a locality also deletes any metrics for the locality.
#curl -u test:test -X DELETE "http://localhost:8080/field/locality?localityID=ahititi"

# Deleting a site removes the code tag from any metrics.
#curl -u test:test -X DELETE "http://localhost:8080/field/site?code=TAUP"

# Deleting a model removes the model and any metrics for it
#curl -u test:test -X DELETE "http://localhost:8080/field/model?modelID=Trimble+NetR9"

# Delete all metrics for a model at a locality
#curl -u test:test -X DELETE "http://localhost:8080/field/metric?localityID=ahititi&modelID=Trimble+NetR9"