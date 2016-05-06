-- mtr schema for objects shared in field, data, and app schemas.
CREATE SCHEMA mtr;

CREATE TABLE mtr.tag (
	tagPK SERIAL PRIMARY KEY,
	tag TEXT NOT NULL UNIQUE
);