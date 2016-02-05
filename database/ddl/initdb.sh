#!/bin/bash

# script for initializing the db in the postgres Docker container.

export PGUSER=postgres

cd /docker-entrypoint-initdb.d

psql  -d postgres < /docker-entrypoint-initdb.d/create-users.ddl
psql  -d postgres < /docker-entrypoint-initdb.d/create-db.ddl
psql  -d mtr -c 'create extension postgis;'
psql  --quiet mtr < /docker-entrypoint-initdb.d/field-schema.ddl
psql  --quiet mtr < /docker-entrypoint-initdb.d/user-permissions.ddl
