CREATE SCHEMA field;

CREATE TABLE field.locality (
	localityPK SERIAL PRIMARY KEY,
	localityID TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	latitude              NUMERIC(8,5) NOT NULL,
    	longitude             NUMERIC(8,5) NOT NULL,
	geom GEOGRAPHY(POINT, 4326) NOT NULL -- added via locality_geom_trigger
);

CREATE FUNCTION field.locality_geom() 
RETURNS  TRIGGER AS 
$$
BEGIN 
NEW.geom = ST_GeogFromWKB(st_AsEWKB(st_setsrid(st_makepoint(NEW.longitude, NEW.latitude), 4326)));
RETURN NEW;  END; 
$$
LANGUAGE plpgsql;

CREATE TRIGGER locality_geom_trigger BEFORE INSERT OR UPDATE ON field.locality
  FOR EACH ROW EXECUTE PROCEDURE field.locality_geom();

CREATE TABLE field.source (
	sourcePK SMALLSERIAL PRIMARY KEY,
	sourceID TEXT NOT NULL UNIQUE
);

CREATE TABLE field.type (
	typePK SMALLINT PRIMARY KEY,
	typeID TEXT NOT NULL UNIQUE,
	description TEXT NOT NULL,
	unit TEXT NOT NULL
);

INSERT INTO field.type(typePK, typeID, description, unit) VALUES(1, 'voltage', 'voltage', 'mV'); 
INSERT INTO field.type(typePK, typeID, description, unit) VALUES(2, 'clock', 'clock quality', 'c'); 
INSERT INTO field.type(typePK, typeID, description, unit) VALUES(3, 'satellites', 'number of statellites tracked', 'n'); 

CREATE TABLE field.metric (
	localityPK INTEGER REFERENCES field.locality(localityPK) ON DELETE CASCADE NOT NULL,
	sourcePK SMALLINT REFERENCES field.source(sourcePK) ON DELETE CASCADE NOT NULL,
	typePK SMALLINT REFERENCES field.type(typePK) ON DELETE CASCADE NOT NULL, 
	time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	value INTEGER NOT NULL,
	PRIMARY KEY(localityPK, sourcePK, typePK, time)
);

CREATE INDEX on field.metric (localityPK);
CREATE INDEX on field.metric (sourcePK);
CREATE INDEX on field.metric (typePK);

CREATE TABLE field.threshold (
	localityPK INTEGER REFERENCES field.locality(localityPK) ON DELETE CASCADE NOT NULL,
	sourcePK SMALLINT REFERENCES field.source(sourcePK) ON DELETE CASCADE NOT NULL,
	typePK SMALLINT REFERENCES field.type(typePK) ON DELETE CASCADE NOT NULL, 
	min INTEGER NOT NULL,
	max INTEGER NOT NULL,
	PRIMARY KEY(localityPK, sourcePK, typePK)
);

CREATE INDEX on field.threshold (localityPK);
CREATE INDEX on field.threshold (sourcePK);
CREATE INDEX on field.threshold (typePK);

CREATE TABLE field.tag (
	tagPK SERIAL PRIMARY KEY,
	tag TEXT NOT NULL UNIQUE
);

CREATE TABLE field.metric_tag(
	localityPK INTEGER REFERENCES field.locality(localityPK) ON DELETE CASCADE NOT NULL,
	sourcePK SMALLINT REFERENCES field.source(sourcePK) ON DELETE CASCADE NOT NULL,
	typePK SMALLINT REFERENCES field.type(typePK) ON DELETE CASCADE NOT NULL, 
	tagPK INTEGER REFERENCES field.tag(tagPK)  ON DELETE CASCADE NOT NULL,
	PRIMARY KEY(localityPK, sourcePK, typePK, tagPK)
);

CREATE INDEX on field.metric_tag (localityPK);
CREATE INDEX on field.metric_tag (sourcePK);
CREATE INDEX on field.metric_tag (typePK);
CREATE INDEX on field.metric_tag (tagPK);


-- field.metric_summary is a materialized view of the metrics for each localityPK, sourcePK, typePK.  
-- it is currently the latest value.  This could be changed.  The time value can be used to see if
-- the metric has stopped sending (purple).
-- The view is refreshed using:
--
-- REFRESH MATERIALIZED VIEW CONCURRENTLY field.metric_summary;
--
-- The data is stale until refreshed, if this causes issues then use an eagerly materialized view using triggers etc.
-- The user that will refresh the view must own it.
CREATE MATERIALIZED VIEW field.metric_summary 
AS SELECT localityPK, sourcePK, typePK, time, value 
FROM 
(SELECT localityPK, sourcePK, typePK, time, value, rank() 
	OVER ( PARTITION BY localityPK, sourcePK, typePK ORDER BY time DESC) FROM field.metric) s 
WHERE rank = 1;

-- UNIQUE index is needed for refresh CONCURRENTLY
CREATE UNIQUE INDEX on field.metric_summary (localityPK, sourcePK, typePk, time);