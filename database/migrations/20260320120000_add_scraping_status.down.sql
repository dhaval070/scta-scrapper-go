ALTER TABLE sites_config
drop COLUMN scraping_status;

ALTER TABLE sites_config
drop COLUMN scraping_started_at;

ALTER TABLE sites_config
drop COLUMN scraping_error;
