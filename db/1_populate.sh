#!/bin/bash -ex

cd `dirname $0`

sqlite3 db.sqlite3 <<EOF
  DROP TABLE IF EXISTS browsers;
  CREATE TABLE browsers (
    id               INTEGER PRIMARY KEY NOT NULL,
    token            TEXT NOT NULL,
    user_agent       TEXT NOT NULL,
    accept           TEXT NOT NULL,
    accept_encoding  TEXT NOT NULL,
    accept_language  TEXT NOT NULL,
    referer          TEXT NOT NULL,
    user_id          INTEGER,
    created_at       INTEGER NOT NULL,
    last_seen_at     INTEGER NOT NULL
  );
  CREATE UNIQUE INDEX idx_browsers_token ON browsers(token);

  DROP TABLE IF EXISTS logs;
  CREATE TABLE logs (
    id                 INTEGER PRIMARY KEY NOT NULL,
    browser_id         INTEGER NOT NULL,
    id_on_client       INTEGER NOT NULL,
    time_on_client     REAL NOT NULL,
    message            TEXT NOT NULL,
    error_name         TEXT,
    error_message      TEXT,
    error_stack        TEXT,
    other_details_json TEXT
  );

  DROP TABLE IF EXISTS recordings;
  CREATE TABLE recordings (
    id                    INTEGER PRIMARY KEY NOT NULL,
    user_id               INTEGER NOT NULL,
    id_on_client          INTEGER NOT NULL,
    recorded_at_on_client REAL NOT NULL,
    uploaded_at           INTEGER NOT NULL,
    filename              TEXT NOT NULL,
    mime_type             TEXT NOT NULL,
    size                  INTEGER NOT NULL,
    metadata_json         TEXT NOT NULL,
    transcript            TEXT
  );

  DROP TABLE IF EXISTS requests;
  CREATE TABLE requests (
    id           INTEGER PRIMARY KEY NOT NULL,
    browser_id   INTEGER,
    http_version TEXT NOT NULL,
    tls_protocol TEXT,
    tls_cipher   TEXT,
    received_at  TEXT NOT NULL,
    remote_addr  TEXT NOT NULL,
    method       TEXT NOT NULL,
    path         TEXT NOT NULL,
    duration_ms  INTEGER NOT NULL,
    status_code  INTEGER NOT NULL,
    size         INTEGER NOT NULL,
    error_stack  TEXT
  );

  DROP TABLE IF EXISTS users;
  CREATE TABLE users (
    id           INTEGER PRIMARY KEY NOT NULL,
    token        TEXT NOT NULL,
    created_at   INTEGER NOT NULL,
    last_seen_at INTEGER NOT NULL
  );
  CREATE UNIQUE INDEX idx_users_token ON users(token);
EOF
