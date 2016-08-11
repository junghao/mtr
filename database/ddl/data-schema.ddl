CREATE SCHEMA data;

CREATE TABLE data.site (
  sitePK SMALLSERIAL PRIMARY KEY,
  siteID TEXT NOT NULL UNIQUE,
  latitude              NUMERIC(8,5) NOT NULL,
  longitude             NUMERIC(8,5) NOT NULL,
  geom GEOGRAPHY(POINT, 4326) NOT NULL -- added via site_geom_trigger
);

CREATE FUNCTION data.site_geom()
  RETURNS  TRIGGER AS
$$
BEGIN
  NEW.geom = ST_GeogFromWKB(st_AsEWKB(st_setsrid(st_makepoint(NEW.longitude, NEW.latitude), 4326)));
  RETURN NEW;  END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER site_geom_trigger BEFORE INSERT OR UPDATE ON data.site
FOR EACH ROW EXECUTE PROCEDURE data.site_geom();

-- metrics are sent as ints in measurement 'unit'.
-- they are scaled for display with 'scale'.
-- 'display' is the unit to display after scaling.
CREATE TABLE data.type (
  typePK SMALLINT PRIMARY KEY,
  typeID TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL,
  unit TEXT NOT NULL,
  scale NUMERIC NOT NULL,
  display TEXT NOT NULL
);

INSERT INTO data.type(typePK, typeID, description, unit, scale, display) VALUES(1, 'latency.strong', 'latency strong motion data', 'ms', 1.0, 'ms');
INSERT INTO data.type(typePK, typeID, description, unit, scale, display) VALUES(2, 'latency.weak', 'latency weak motion data', 'ms', 1.0, 'ms');
INSERT INTO data.type(typePK, typeID, description, unit, scale, display) VALUES(3, 'latency.gnss.1hz', 'latency GNSS 1Hz data', 'ms', 1.0, 'ms');
INSERT INTO data.type(typePK, typeID, description, unit, scale, display) VALUES(4, 'latency.tsunami', 'latency tsunami data', 'ms', 1.0, 'ms');
INSERT INTO data.type(typePK, typeID, description, unit, scale, display) VALUES(5, 'latency.files.gnss', 'latency files data', 'ms', 1.0, 'ms');

CREATE TABLE data.latency (
  sitePK INTEGER REFERENCES data.site(sitePK) ON DELETE CASCADE NOT NULL,
  typePK SMALLINT REFERENCES data.type(typePK) ON DELETE CASCADE NOT NULL,
  rate_limit BIGINT NOT NULL,
  time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
  mean INTEGER NOT NULL,
  min INTEGER NOT NULL,
  max INTEGER NOT NULL,
  fifty INTEGER NOT NULL,
  ninety INTEGER NOT NULL,
  PRIMARY KEY(sitePK, typePK, rate_limit)
);

CREATE INDEX ON data.latency (time);

CREATE TABLE data.latency_summary (
  sitePK INTEGER REFERENCES data.site(sitePK) ON DELETE CASCADE NOT NULL,
  typePK SMALLINT REFERENCES data.type(typePK) ON DELETE CASCADE NOT NULL,
  time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
  mean INTEGER NOT NULL,
  min INTEGER NOT NULL,
  max INTEGER NOT NULL,
  fifty INTEGER NOT NULL,
  ninety INTEGER NOT NULL,
  PRIMARY KEY(sitePK, typePK)
);

CREATE TABLE data.latency_threshold (
  sitePK SMALLINT REFERENCES data.site(sitePK) ON DELETE CASCADE NOT NULL,
  typePK SMALLINT REFERENCES data.type(typePK) ON DELETE CASCADE NOT NULL,
  lower INTEGER NOT NULL,
  upper INTEGER NOT NULL,
  PRIMARY KEY(sitePK, typePK)
);

CREATE TABLE data.latency_tag(
  sitePK SMALLINT REFERENCES data.site(sitePK) ON DELETE CASCADE NOT NULL,
  typePK SMALLINT REFERENCES data.type(typePK) ON DELETE CASCADE NOT NULL,
  tagPK INTEGER REFERENCES mtr.tag(tagPK) ON DELETE CASCADE NOT NULL,
  PRIMARY KEY(sitePK, typePK, tagPK)
);

-- expected is the expected counts in a 24 hour period.
CREATE TABLE data.completeness_type (
  typePK SMALLINT PRIMARY KEY,
  typeID TEXT NOT NULL UNIQUE,
  expected INTEGER NOT NULL
);

INSERT INTO data.completeness_type(typePK, typeID, expected) VALUES(100, 'completeness.gnss.1hz', 86400);

CREATE TABLE data.completeness (
  sitePK INTEGER REFERENCES data.site(sitePK) ON DELETE CASCADE NOT NULL,
  typePK SMALLINT REFERENCES data.completeness_type(typePK) ON DELETE CASCADE NOT NULL,
  rate_limit BIGINT NOT NULL,
  time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
  count INTEGER NOT NULL,
  PRIMARY KEY(sitePK, typePK, rate_limit)
);

CREATE INDEX ON data.completeness (time);

CREATE TABLE data.completeness_summary (
  sitePK INTEGER REFERENCES data.site(sitePK) ON DELETE CASCADE NOT NULL,
  typePK SMALLINT REFERENCES data.completeness_type(typePK) ON DELETE CASCADE NOT NULL,
  time TIMESTAMP(0) WITH TIME ZONE NOT NULL,
  count INTEGER NOT NULL,
  PRIMARY KEY(sitePK, typePK)
);

CREATE TABLE data.completeness_tag(
  sitePK SMALLINT REFERENCES data.site(sitePK) ON DELETE CASCADE NOT NULL,
  typePK SMALLINT REFERENCES data.completeness_type(typePK) ON DELETE CASCADE NOT NULL,
  tagPK INTEGER REFERENCES mtr.tag(tagPK) ON DELETE CASCADE NOT NULL,
  PRIMARY KEY(sitePK, typePK, tagPK)
);