CREATE TABLE quotedb (
  channel STRING NOT NULL,
  created_at INTEGER,
  quote STRING NOT NULL,

  UNIQUE(channel, created_at)
);
