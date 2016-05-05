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

CREATE TABLE data.type (
  typePK SMALLINT PRIMARY KEY,
  typeID TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL,
  unit TEXT NOT NULL
);

--  These must also be added to mtr-api/data_type.go
INSERT INTO data.type(typePK, typeID, description, unit) VALUES(1, 'latency.strong', 'latency strong motion data', 'ms');
INSERT INTO data.type(typePK, typeID, description, unit) VALUES(2, 'latency.weak', 'latency weak motion data', 'ms');
INSERT INTO data.type(typePK, typeID, description, unit) VALUES(3, 'latency.gnss.1hz', 'latency GNSS 1Hz data', 'ms');
INSERT INTO data.type(typePK, typeID, description, unit) VALUES(4, 'latency.tsunami', 'latency tsunami data', 'ms');

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

CREATE INDEX on data.latency (sitePK);
CREATE INDEX on data.latency (typePK);
CREATE INDEX on data.latency (time);

CREATE TABLE data.latency_threshold (
  sitePK SMALLINT REFERENCES data.site(sitePK) ON DELETE CASCADE NOT NULL,
  typePK SMALLINT REFERENCES data.type(typePK) ON DELETE CASCADE NOT NULL,
  lower INTEGER NOT NULL,
  upper INTEGER NOT NULL,
  PRIMARY KEY(sitePK, typePK)
);

CREATE INDEX on data.latency_threshold (sitePK);
CREATE INDEX on data.latency_threshold (typePK);