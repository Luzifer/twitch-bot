CREATE TABLE punish_levels (
  key STRING NOT NULL PRIMARY KEY,
  last_level INTEGER,
  executed INTEGER, -- time.Time
  cooldown INTEGER -- time.Duration
);
