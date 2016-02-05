CREATE SCHEMA field;

CREATE TABLE field.locality (
	localityPK SERIAL PRIMARY KEY,
	localityID TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	latitude              NUMERIC NOT NULL,
    	longitude             NUMERIC NOT NULL,
	geom GEOGRAPHY(POINT, 4326) NOT NULL, -- added via locality_geom_trigger
    	geom3857 GEOMETRY(POINT, 3857) NOT NULL -- added via locality_geom_trigger
);

CREATE FUNCTION field.locality_geom() 
RETURNS  TRIGGER AS 
$$
BEGIN 
NEW.geom = ST_GeogFromWKB(st_AsEWKB(st_setsrid(st_makepoint(NEW.longitude, NEW.latitude), 4326)));
NEW.geom3857 = ST_Transform(NEW.geom::geometry, 3857);
RETURN NEW;  END; 
$$
LANGUAGE plpgsql;

CREATE TRIGGER locality_geom_trigger BEFORE INSERT OR UPDATE ON field.locality
  FOR EACH ROW EXECUTE PROCEDURE field.locality_geom();

CREATE TABLE field.model (
	modelPK SMALLSERIAL PRIMARY KEY,
	modelID TEXT NOT NULL UNIQUE
);

CREATE TABLE field.metricType (
       metricTypePK SMALLINT PRIMARY KEY,
       metricTypeID TEXT NOT NULL UNIQUE,
       description TEXT NOT NULL,
       unit TEXT NOT NULL
);

-- Field metrics type additions also need making to internal/field_types.go
INSERT INTO field.metricType(metricTypePK, metricTypeID, description, unit) VALUES(1, 'voltage', 'voltage', 'mV'); 
INSERT INTO field.metricType(metricTypePK, metricTypeID, description, unit) VALUES(2, 'clock', 'clock quality', 'c'); 
INSERT INTO field.metricType(metricTypePK, metricTypeID, description, unit) VALUES(3, 'satellites', 'number of statellites tracked', 'n'); 

CREATE TABLE field.site (
	sitePK SMALLSERIAL PRIMARY KEY,
	code TEXT NOT NULL UNIQUE
);

INSERT INTO field.site(sitePK, code) VALUES(0, 'NO-CODE');

CREATE TABLE field.metric (
	localityPK INTEGER REFERENCES field.locality(localityPK) ON DELETE CASCADE NOT NULL,
	modelPK SMALLINT REFERENCES field.model(modelPK) ON DELETE CASCADE NOT NULL,
	sitePK SMALLINT  DEFAULT 0 REFERENCES field.site(sitePK) ON DELETE SET DEFAULT NOT NULL,
	metricTypePK SMALLINT REFERENCES field.metricType(metricTypePK) ON DELETE CASCADE NOT NULL, 
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	value INTEGER NOT NULL,
	PRIMARY KEY(localityPK, modelPK, metricTypePK, time)
);

CREATE INDEX on field.metric (localityPK);
CREATE INDEX on field.metric (modelPK);
CREATE INDEX on field.metric (metricTypePK);
CREATE INDEX on field.metric (sitePK);

