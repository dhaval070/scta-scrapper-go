ALTER TABLE `mhr_locations`
ADD `livebarn_location_id` INT NULL DEFAULT '0' AFTER `livebarn_installed`,
ADD `livebarn_surface_id` INT NULL DEFAULT '0' AFTER `livebarn_location_id`,
ADD INDEX `idx_lb_loc_id` (`livebarn_location_id`),
ADD INDEX `idx_lb_surface_id` (`livebarn_surface_id`);
