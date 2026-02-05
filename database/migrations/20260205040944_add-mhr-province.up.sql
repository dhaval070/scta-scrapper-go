ALTER TABLE `mhr_locations` ADD `home_teams` JSON NULL AFTER `livebarn_surface_id`;
ALTER TABLE `mhr_locations` ADD `province` VARCHAR(64) NOT NULL AFTER `home_teams`, ADD INDEX `idx_province` (`province`);
