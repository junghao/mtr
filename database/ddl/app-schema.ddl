CREATE SCHEMA app;

CREATE TABLE app.application (
	applicationPK SMALLSERIAL PRIMARY KEY,
	applicationID TEXT NOT NULL UNIQUE
);

-- instance.id should uniquely identify a running instance of an application.
-- e.g., uuid or host or location.
CREATE TABLE app.instance (
	instancePK SMALLSERIAL PRIMARY KEY,
	instanceID TEXT NOT NULL UNIQUE
);

-- source for timer e.g., a method or function name.
CREATE TABLE app.source (
	sourcePK SERIAL PRIMARY KEY,
	sourceID TEXT NOT NULL UNIQUE
);

CREATE TABLE app.type (
       typePK SMALLINT PRIMARY KEY,
       typeID TEXT NOT NULL UNIQUE,
       description TEXT NOT NULL,
       unit TEXT NOT NULL
);

CREATE TABLE app.counter (
	applicationPK SMALLINT REFERENCES app.application(applicationPK) NOT NULL,
	instancePK SMALLINT REFERENCES app.instance(instancePK) NOT NULL,
	typePK SMALLINT REFERENCES app.type(typePK) NOT NULL, 
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	count INTEGER NOT NULL,
	PRIMARY KEY(applicationPK, instancePK, typePK, time)
);

CREATE INDEX on app.counter (applicationPK);
CREATE INDEX on app.counter (instancePK);
CREATE INDEX on app.counter (time);

CREATE TABLE app.timer (
	applicationPK SMALLINT REFERENCES app.application(applicationPK) NOT NULL,
	instancePK SMALLINT REFERENCES app.instance(instancePK) NOT NULL,
	sourcePK INTEGER REFERENCES app.source(sourcePK) NOT NULL,
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	total INTEGER NOT NULL,
	count INTEGER NOT NULL, 
	PRIMARY KEY(applicationPK, instancePK, sourcePK, time)
);

CREATE INDEX on app.timer (applicationPK);
CREATE INDEX on app.timer (instancePK);
CREATE INDEX on app.timer (sourcePK);
CREATE INDEX on app.timer (time);

CREATE TABLE app.metric (
	applicationPK SMALLINT REFERENCES app.application(applicationPK) NOT NULL,
	instancePK SMALLINT REFERENCES app.instance(instancePK) NOT NULL,
	typePK SMALLINT REFERENCES app.type(typePK) NOT NULL, 
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	value BIGINT NOT NULL,
	PRIMARY KEY(applicationPK, instancePK, typePK, time)
);

CREATE INDEX on app.metric (applicationPK);
CREATE INDEX on app.metric (instancePK);
CREATE INDEX on app.metric (time);

--- HTTP Requests
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(1, 'Requests', 'Requests', 'n'); 

--- HTTP Status codes 100 - 999
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(200, 'StatusOK', 'OK', 'n'); 
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(400, 'StatusBadRequest', 'Bad Request', 'n'); 
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(404, 'StatusNotFound', 'Not Found', 'n'); 
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(500, 'StatusInternalServerError', 'Internal Server Error', 'n'); 
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(503, 'StatusServiceUnavailable', 'Service Unavailable', 'n'); 

--- MemStats
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(1000, 'MemSys', 'bytes obtained from system', 'bytes'); 
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(1001, 'MemHeapAlloc', 'bytes allocated and not yet freed', 'bytes'); 
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(1002, 'MemHeapSys', 'bytes obtained from system', 'bytes'); 
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(1003, 'MemHeapObjects', 'total number of allocated objects', 'n'); 

--- Other runtime stats
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(1100, 'Routines', 'number of routines that currently exist', 'n'); 