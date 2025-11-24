UPDATE events 
SET site = CONCAT('gs_', SUBSTRING(site, 1, LENGTH(site) - 3))
WHERE site LIKE '%\_gs';

UPDATE gamesheet_seasons 
SET site = CONCAT('gs_', SUBSTRING(site, 1, LENGTH(site) - 3))
WHERE site LIKE '%\_gs';

UPDATE sites 
SET site = CONCAT('gs_', SUBSTRING(site, 1, LENGTH(site) - 3))
WHERE site LIKE '%\_gs';

UPDATE sites_config 
SET site_name = CONCAT('gs_', SUBSTRING(site_name, 1, LENGTH(site_name) - 3))
WHERE site_name LIKE '%\_gs';

UPDATE sites_locations 
SET site = CONCAT('gs_', SUBSTRING(site, 1, LENGTH(site) - 3))
WHERE site LIKE '%\_gs';

