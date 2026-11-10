#!/bin/sh
set -e

psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d postgres \
  -c "CREATE DATABASE wallet_test_db"

psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d wallet_test_db \
  -f /docker-entrypoint-initdb.d/01_create_tables.sql