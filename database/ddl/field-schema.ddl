CREATE SCHEMA field;

CREATE TABLE field.model (
	modelPK SMALLSERIAL PRIMARY KEY,
	modelID TEXT NOT NULL UNIQUE
);

CREATE TABLE field.device (
	devicePK SMALLSERIAL PRIMARY KEY,
	deviceID TEXT NOT NULL UNIQUE,
	modelPK SMALLINT REFERENCES field.model(modelPK) ON DELETE CASCADE NOT NULL,
	latitude              NUMERIC(8,5) NOT NULL,
	longitude             NUMERIC(8,5) NOT NULL,
	geom GEOGRAPHY(POINT, 4326) NOT NULL -- added via device_geom_trigger
);

CREATE FUNCTION field.device_geom() 
RETURNS  TRIGGER AS 
$$
BEGIN 
NEW.geom = ST_GeogFromWKB(st_AsEWKB(st_setsrid(st_makepoint(NEW.longitude, NEW.latitude), 4326)));
RETURN NEW;  END; 
$$
LANGUAGE plpgsql;

CREATE TRIGGER device_geom_trigger BEFORE INSERT OR UPDATE ON field.device
FOR EACH ROW EXECUTE PROCEDURE field.device_geom();

-- metrics are sent as ints in measurement 'unit'.
-- they are scaled for display with 'scale'.
-- 'display' is the unit to display after scaling.
CREATE TABLE field.type (
	typePK SMALLINT PRIMARY KEY,
	typeID TEXT NOT NULL UNIQUE,
	description TEXT NOT NULL,
	unit TEXT NOT NULL,
	scale NUMERIC NOT NULL,
	display TEXT NOT NULL
);

-- If types are added the application server will need restarting.
INSERT INTO field.type(typePK, typeID, description, unit, scale, display) VALUES(1, 'voltage', 'voltage', 'mV', 0.001, 'V');
INSERT INTO field.type(typePK, typeID, description, unit, scale, display) VALUES(2, 'clock', 'clock quality', '%', 1.0, '%');
INSERT INTO field.type(typePK, typeID, description, unit, scale, display) VALUES(3, 'satellites', 'number of satellites tracked', 'n', 1.0, 'n');
INSERT INTO field.type(typePK, typeID, description, unit, scale, display) VALUES(4, 'conn', 'end to end connectivity', 'us', 0.001, 'ms');
INSERT INTO field.type(typePK, typeID, description, unit, scale, display) VALUES(5, 'ping', 'ping', 'us', 0.001, 'ms');

INSERT INTO field.type(typePK, typeID, description, unit, scale, display) VALUES(6, 'disk.hd1', 'disk hd1', '%', 1.0, '%');
INSERT INTO field.type(typePK, typeID, description, unit, scale, display) VALUES(7, 'disk.hd2', 'disk hd1', '%', 1.0, '%');
INSERT INTO field.type(typePK, typeID, description, unit, scale, display) VALUES(8, 'disk.hd3', 'disk hd1', '%', 1.0, '%');
INSERT INTO field.type(typePK, typeID, description, unit, scale, display) VALUES(9, 'disk.hd4', 'disk hd1', '%', 1.0, '%');

INSERT INTO field.type(typePK, typeID, description, unit, scale, display) VALUES(10, 'centre', 'centre', 'mV', 1.0, 'mV');

INSERT INTO field.type(typePK, typeID, description, unit, scale, display) VALUES(11, 'rf.signal', 'rf signal', 'dB', 1.0, 'db');
INSERT INTO field.type(typePK, typeID, description, unit, scale, display) VALUES(12, 'rf.noise', 'rf signal', 'dB', 1.0, 'db');

CREATE TABLE field.metric (
	devicePK SMALLINT REFERENCES field.device(devicePK) ON DELETE CASCADE NOT NULL,
	typePK SMALLINT REFERENCES field.type(typePK) ON DELETE CASCADE NOT NULL,
	rate_limit BIGINT NOT NULL,
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	value INTEGER NOT NULL,
	PRIMARY KEY(devicePK, typePK, rate_limit)
);

CREATE INDEX on field.metric (devicePK);
CREATE INDEX on field.metric (typePK);

CREATE TABLE field.threshold (
	devicePK SMALLINT REFERENCES field.device(devicePK) ON DELETE CASCADE NOT NULL,
	typePK SMALLINT REFERENCES field.type(typePK) ON DELETE CASCADE NOT NULL, 
	lower INTEGER NOT NULL,
	upper INTEGER NOT NULL,
	PRIMARY KEY(devicePK, typePK)
);

CREATE INDEX on field.threshold (devicePK);
CREATE INDEX on field.threshold (typePK);

CREATE TABLE field.metric_tag(
	devicePK SMALLINT REFERENCES field.device(devicePK) ON DELETE CASCADE NOT NULL,
	typePK SMALLINT REFERENCES field.type(typePK) ON DELETE CASCADE NOT NULL, 
	tagPK INTEGER REFERENCES mtr.tag(tagPK) ON DELETE CASCADE NOT NULL,
	PRIMARY KEY(devicePK, typePK, tagPK)
);

CREATE INDEX on field.metric_tag (devicePK);
CREATE INDEX on field.metric_tag (typePK);
CREATE INDEX on field.metric_tag (tagPK);
