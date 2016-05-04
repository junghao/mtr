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


