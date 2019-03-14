#!/bin/bash
set -e

# create user
psql -v ON_ERROR_STOP=1 --username "postgres" postgres <<-EOSQL
    CREATE ROLE business LOGIN ENCRYPTED PASSWORD 'md5c62bf6b9e93f660a9f70b84048d45fda' VALID UNTIL 'infinity';
    CREATE DATABASE business WITH ENCODING='UTF8' OWNER=business CONNECTION LIMIT=-1;
EOSQL