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

CREATE TABLE app.counter_minute (
	applicationPK SMALLINT REFERENCES app.application(applicationPK) NOT NULL,
	typePK SMALLINT REFERENCES app.type(typePK) NOT NULL, 
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	count INTEGER NOT NULL,
	PRIMARY KEY(applicationPK, typePK, time)
);

CREATE INDEX on app.counter_minute (applicationPK);
CREATE INDEX on app.counter_minute (time);

CREATE TABLE app.counter_hour (
	applicationPK SMALLINT REFERENCES app.application(applicationPK) NOT NULL,
	typePK SMALLINT REFERENCES app.type(typePK) NOT NULL, 
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	count INTEGER NOT NULL,
	PRIMARY KEY(applicationPK, typePK, time)
);

CREATE INDEX on app.counter_hour (applicationPK);
CREATE INDEX on app.counter_hour (time);

CREATE TABLE app.timer_minute (
	applicationPK SMALLINT REFERENCES app.application(applicationPK) NOT NULL,
	sourcePK INTEGER REFERENCES app.source(sourcePK) NOT NULL,
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	avg INTEGER NOT NULL,
	n INTEGER NOT NULL, 
	PRIMARY KEY(applicationPK, sourcePK, time)
);

CREATE INDEX on app.timer_minute (applicationPK);
CREATE INDEX on app.timer_minute (sourcePK);
CREATE INDEX on app.timer_minute (time);

CREATE TABLE app.timer_hour (
	applicationPK SMALLINT REFERENCES app.application(applicationPK) NOT NULL,
	sourcePK INTEGER REFERENCES app.source(sourcePK) NOT NULL,
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	avg INTEGER NOT NULL,
	n INTEGER NOT NULL, 
	PRIMARY KEY(applicationPK, sourcePK, time)
);

CREATE INDEX on app.timer_hour (applicationPK);
CREATE INDEX on app.timer_hour (sourcePK);
CREATE INDEX on app.timer_hour (time);

CREATE TABLE app.metric_minute (
	applicationPK SMALLINT REFERENCES app.application(applicationPK) NOT NULL,
	instancePK SMALLINT REFERENCES app.instance(instancePK) NOT NULL,
	typePK SMALLINT REFERENCES app.type(typePK) NOT NULL, 
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	avg BIGINT NOT NULL,
	n INTEGER NOT NULL,
	PRIMARY KEY(applicationPK, instancePK, typePK, time)
);

CREATE INDEX on app.metric_minute (applicationPK);
CREATE INDEX on app.metric_minute (instancePK);
CREATE INDEX on app.metric_minute (time);

CREATE TABLE app.metric_hour (
	applicationPK SMALLINT REFERENCES app.application(applicationPK) NOT NULL,
	instancePK SMALLINT REFERENCES app.instance(instancePK) NOT NULL,
	typePK SMALLINT REFERENCES app.type(typePK) NOT NULL, 
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	avg BIGINT NOT NULL,
	n INTEGER NOT NULL,
	PRIMARY KEY(applicationPK, instancePK, typePK, time)
);

CREATE INDEX on app.metric_hour (applicationPK);
CREATE INDEX on app.metric_hour (instancePK);
CREATE INDEX on app.metric_hour (time);

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