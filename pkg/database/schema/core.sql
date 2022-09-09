-- Core database structure, to be applied before any migration

CREATE TABLE IF NOT EXISTS core_kv (
  key STRING NOT NULL PRIMARY KEY,
  value STRING
);
