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

  DROP TABLE IF EXISTS capabilities;
  CREATE TABLE capabilities (
    id                            INTEGER PRIMARY KEY NOT NULL,
    request_id                    INTEGER NOT NULL,
    created_at                    TEXT NOT NULL,
    navigator_app_code_name       TEXT,
    navigator_app_name            TEXT,
    navigator_app_version         TEXT,
    navigator_cookie_enabled      TEXT,
    navigator_language            TEXT,
    navigator_languages           TEXT,
    navigator_platform            TEXT,
    navigator_oscpu               TEXT,
    navigator_user_agent          TEXT,
    navigator_vendor              TEXT,
    navigator_vendor_sub          TEXT,
    screen_width                  TEXT,
    screen_height                 TEXT,
    window_inner_width            TEXT,
    window_inner_height           TEXT,
    doc_body_client_width         TEXT,
    doc_body_client_height        TEXT,
    doc_element_client_width      TEXT,
    doc_element_client_height     TEXT,
    window_screen_avail_width     TEXT,
    window_screen_avail_height    TEXT,
    window_device_pixel_ratio     TEXT,
    has_on_touch_start            TEXT
  );

EOF
