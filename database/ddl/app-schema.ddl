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
	applicationPK SMALLINT REFERENCES app.application(applicationPK) ON DELETE CASCADE NOT NULL,
	instancePK SMALLINT REFERENCES app.instance(instancePK) ON DELETE CASCADE NOT NULL,
	typePK SMALLINT REFERENCES app.type(typePK) ON DELETE CASCADE NOT NULL,
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	count INTEGER NOT NULL,
	PRIMARY KEY(applicationPK, instancePK, typePK, time)
);

CREATE INDEX ON app.counter (time);

CREATE TABLE app.timer (
	applicationPK SMALLINT REFERENCES app.application(applicationPK) ON DELETE CASCADE NOT NULL,
	instancePK SMALLINT REFERENCES app.instance(instancePK) ON DELETE CASCADE NOT NULL,
	sourcePK INTEGER REFERENCES app.source(sourcePK) ON DELETE CASCADE NOT NULL,
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	average INTEGER NOT NULL,
	count INTEGER NOT NULL,
	fifty INTEGER NOT NULL,
	ninety INTEGER NOT NULL,
	PRIMARY KEY(applicationPK, instancePK, sourcePK, time)
);

CREATE INDEX ON app.timer (time);

CREATE TABLE app.metric (
	applicationPK SMALLINT REFERENCES app.application(applicationPK) ON DELETE CASCADE NOT NULL,
	instancePK SMALLINT REFERENCES app.instance(instancePK) ON DELETE CASCADE NOT NULL,
	typePK SMALLINT REFERENCES app.type(typePK) ON DELETE CASCADE NOT NULL,
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	value BIGINT NOT NULL,
	PRIMARY KEY(applicationPK, instancePK, typePK, time)
);

CREATE INDEX ON app.metric (time);

--- HTTP Requests
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(1, 'Requests', 'Requests', 'n'); 

--- HTTP Status codes 100 - 999
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(200, 'StatusOK', 'OK', 'n'); 
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(400, 'StatusBadRequest', 'Bad Request', 'n'); 
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(401, 'StatusUnauthorized', 'Unauthorized', 'n'); 
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

--- Message counters
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(1201, 'MsgRx', 'messages received', 'n'); 
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(1202, 'MsgTx', 'messages transmitted', 'n'); 
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(1203, 'MsgProc', 'messages processed', 'n'); 
INSERT INTO app.type(typePK, typeID, description, unit) VALUES(1204, 'MsgErr', 'messages error', 'n'); 