ALTER TABLE sites_config
ADD COLUMN scraping_status VARCHAR(20) DEFAULT 'idle' COMMENT 'Current scraping status';

ALTER TABLE sites_config
ADD COLUMN scraping_started_at TIMESTAMP NULL COMMENT 'When scraping started';

ALTER TABLE sites_config
ADD COLUMN scraping_error TEXT COMMENT 'Error message if scraping failed';

ALTER TABLE sites_config
ADD CONSTRAINT chk_scraping_status 
    CHECK (scraping_status IN ('idle', 'running', 'completed', 'failed'));

CREATE INDEX idx_scraping_status ON sites_config (scraping_status);
