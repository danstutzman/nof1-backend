#!/bin/bash -ex

cd `dirname $0`

sqlite3 db.sqlite3 <<EOF
  DROP TABLE IF EXISTS requests;
  CREATE TABLE requests (
    id          INTEGER PRIMARY KEY NOT NULL,
    remote_addr TEXT NOT NULL,
    user_agent  TEXT NOT NULL,
    referer     TEXT NOT NULL,
    method      TEXT NOT NULL,
    path        TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    size        INTEGER NOT NULL,
    received_at TEXT NOT NULL
  );
EOF
