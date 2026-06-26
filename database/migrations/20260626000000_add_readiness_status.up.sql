ALTER TABLE sites_config ADD COLUMN readiness_status tinyint NOT NULL DEFAULT 0 COMMENT '0=pending, 1=in progress, 2=ready' AFTER scraping_error;
