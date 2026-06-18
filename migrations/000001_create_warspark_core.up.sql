CREATE TABLE base_layouts (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  th_level INT NOT NULL,
  layout_type VARCHAR(64) NOT NULL,
  style_tags JSON NULL,
  source_type VARCHAR(64) NOT NULL,
  source_url TEXT NULL,
  review_status VARCHAR(64) NOT NULL,
  quality_status VARCHAR(64) NOT NULL,
  visibility VARCHAR(32) NOT NULL DEFAULT 'hidden',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  INDEX idx_base_layouts_public_filter (visibility, th_level, layout_type, review_status, quality_status),
  INDEX idx_base_layouts_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE layout_images (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  layout_id VARCHAR(36) NULL,
  image_url TEXT NOT NULL,
  source_type VARCHAR(64) NOT NULL,
  source_url TEXT NULL,
  width INT NULL,
  height INT NULL,
  image_role VARCHAR(64) NOT NULL,
  review_status VARCHAR(64) NOT NULL,
  quality_status VARCHAR(64) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_layout_images_layout_role (layout_id, image_role),
  CONSTRAINT fk_layout_images_layout FOREIGN KEY (layout_id) REFERENCES base_layouts(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE layout_links (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  layout_id VARCHAR(36) NOT NULL,
  link_type VARCHAR(64) NOT NULL,
  url TEXT NOT NULL,
  link_status VARCHAR(64) NOT NULL,
  source_type VARCHAR(64) NOT NULL,
  source_url TEXT NULL,
  last_checked_at DATETIME(3) NULL,
  last_check_error TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_layout_links_layout_status (layout_id, link_status),
  INDEX idx_layout_links_type (link_type),
  CONSTRAINT fk_layout_links_layout FOREIGN KEY (layout_id) REFERENCES base_layouts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE videos (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  youtube_video_id VARCHAR(64) NOT NULL,
  title VARCHAR(255) NOT NULL,
  channel_name VARCHAR(255) NULL,
  published_at DATETIME(3) NULL,
  source_type VARCHAR(64) NOT NULL,
  source_url TEXT NULL,
  review_status VARCHAR(64) NOT NULL,
  visibility VARCHAR(32) NOT NULL DEFAULT 'hidden',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_videos_youtube_video_id (youtube_video_id),
  INDEX idx_videos_visibility (visibility, review_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE layout_video_matches (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  layout_id VARCHAR(36) NOT NULL,
  video_id VARCHAR(36) NOT NULL,
  timestamp_seconds INT NOT NULL,
  match_group VARCHAR(64) NOT NULL,
  match_type VARCHAR(64) NOT NULL,
  stars INT NULL,
  destruction_percent DECIMAL(5,2) NULL,
  video_tags JSON NULL,
  confidence_score DECIMAL(5,4) NULL,
  review_status VARCHAR(64) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_layout_video_matches_layout_group (layout_id, match_group, match_type),
  INDEX idx_layout_video_matches_video (video_id),
  CONSTRAINT fk_layout_video_matches_layout FOREIGN KEY (layout_id) REFERENCES base_layouts(id) ON DELETE CASCADE,
  CONSTRAINT fk_layout_video_matches_video FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE image_search_jobs (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  uploaded_image_id VARCHAR(36) NULL,
  search_status VARCHAR(64) NOT NULL,
  detected_th INT NULL,
  screenshot_quality VARCHAR(64) NULL,
  buildings_detected INT NULL,
  upload_ip_hash VARCHAR(128) NULL,
  original_retention_until DATETIME(3) NULL,
  processing_mode VARCHAR(64) NOT NULL DEFAULT 'auto',
  error_code VARCHAR(64) NULL,
  error_message TEXT NULL,
  target_context JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  INDEX idx_image_search_jobs_status (search_status, created_at),
  INDEX idx_image_search_jobs_uploaded_image (uploaded_image_id),
  CONSTRAINT fk_image_search_jobs_uploaded_image FOREIGN KEY (uploaded_image_id) REFERENCES layout_images(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE image_search_results (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  search_job_id VARCHAR(36) NOT NULL,
  layout_id VARCHAR(36) NULL,
  rank_order INT NOT NULL,
  match_level VARCHAR(64) NOT NULL,
  confidence_score DECIMAL(5,4) NULL,
  result_reason TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_image_search_results_job_rank (search_job_id, rank_order),
  INDEX idx_image_search_results_layout (layout_id),
  CONSTRAINT fk_image_search_results_job FOREIGN KEY (search_job_id) REFERENCES image_search_jobs(id) ON DELETE CASCADE,
  CONSTRAINT fk_image_search_results_layout FOREIGN KEY (layout_id) REFERENCES base_layouts(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE war_snapshots (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  clan_tag VARCHAR(32) NOT NULL,
  opponent_clan_tag VARCHAR(32) NULL,
  war_state VARCHAR(64) NOT NULL,
  team_size INT NULL,
  clan_stars INT NULL,
  opponent_stars INT NULL,
  clan_destruction DECIMAL(5,2) NULL,
  opponent_destruction DECIMAL(5,2) NULL,
  source_type VARCHAR(64) NOT NULL,
  fetched_at DATETIME(3) NOT NULL,
  api_error_code VARCHAR(64) NULL,
  api_error_message TEXT NULL,
  INDEX idx_war_snapshots_clan_fetched (clan_tag, fetched_at),
  INDEX idx_war_snapshots_state (war_state)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE war_members (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  war_snapshot_id VARCHAR(36) NOT NULL,
  side VARCHAR(32) NOT NULL,
  map_position INT NOT NULL,
  player_tag VARCHAR(32) NULL,
  player_name VARCHAR(255) NOT NULL,
  th_level INT NULL,
  attacks_used INT NULL,
  best_stars_against INT NULL,
  best_destruction_against DECIMAL(5,2) NULL,
  INDEX idx_war_members_snapshot_side_position (war_snapshot_id, side, map_position),
  CONSTRAINT fk_war_members_snapshot FOREIGN KEY (war_snapshot_id) REFERENCES war_snapshots(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE war_targets (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  war_snapshot_id VARCHAR(36) NOT NULL,
  war_member_id VARCHAR(36) NOT NULL,
  search_job_id VARCHAR(36) NULL,
  target_position INT NOT NULL,
  target_name VARCHAR(255) NOT NULL,
  target_th INT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_war_targets_snapshot_position (war_snapshot_id, target_position),
  INDEX idx_war_targets_search_job (search_job_id),
  CONSTRAINT fk_war_targets_snapshot FOREIGN KEY (war_snapshot_id) REFERENCES war_snapshots(id) ON DELETE CASCADE,
  CONSTRAINT fk_war_targets_member FOREIGN KEY (war_member_id) REFERENCES war_members(id) ON DELETE CASCADE,
  CONSTRAINT fk_war_targets_search_job FOREIGN KEY (search_job_id) REFERENCES image_search_jobs(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE import_batches (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  source_type VARCHAR(64) NOT NULL,
  source_url TEXT NULL,
  file_url TEXT NULL,
  import_status VARCHAR(64) NOT NULL,
  total_rows INT NULL,
  success_rows INT NULL,
  failed_rows INT NULL,
  error_summary TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  INDEX idx_import_batches_status (import_status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE admin_audit_logs (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  admin_id VARCHAR(64) NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id VARCHAR(36) NOT NULL,
  action VARCHAR(64) NOT NULL,
  before_snapshot JSON NULL,
  after_snapshot JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_admin_audit_logs_resource (resource_type, resource_id),
  INDEX idx_admin_audit_logs_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
