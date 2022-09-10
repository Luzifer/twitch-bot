CREATE TABLE overlays_events (
  channel STRING NOT NULL,
  created_at INTEGER,
  event_type STRING,
  fields STRING
);

CREATE INDEX overlays_events_sort_idx
  ON overlays_events (channel, created_at DESC);
