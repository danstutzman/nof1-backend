#!/bin/bash -ex

cd `dirname $0`

sqlite3 db.sqlite3 <<EOF
  DROP TABLE IF EXISTS requests;
  CREATE TABLE requests (
    id           INTEGER PRIMARY KEY NOT NULL,
    received_at  TEXT NOT NULL,
    remote_addr  TEXT NOT NULL,
    http_version TEXT NOT NULL,
    tls_protocol TEXT,
    tls_cipher   TEXT,
    user_agent   TEXT NOT NULL,
    referer      TEXT NOT NULL,
    method       TEXT NOT NULL,
    path         TEXT NOT NULL,
    duration_ms  INTEGER NOT NULL,
    status_code  INTEGER NOT NULL,
    size         INTEGER NOT NULL
  );
EOF
