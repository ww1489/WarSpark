CREATE TABLE clan_snapshots (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  clan_tag VARCHAR(32) NOT NULL,
  detail JSON NOT NULL,
  fetched_at DATETIME(3) NOT NULL,
  INDEX idx_clan_snapshots_tag_fetched (clan_tag, fetched_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE player_snapshots (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  player_tag VARCHAR(32) NOT NULL,
  overview JSON NOT NULL,
  fetched_at DATETIME(3) NOT NULL,
  INDEX idx_player_snapshots_tag_fetched (player_tag, fetched_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
