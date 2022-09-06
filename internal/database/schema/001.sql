CREATE TABLE extended_permissions (
  channel STRING NOT NULL PRIMARY KEY,
  access_token STRING,
  refresh_token STRING,
  scopes STRING
);
