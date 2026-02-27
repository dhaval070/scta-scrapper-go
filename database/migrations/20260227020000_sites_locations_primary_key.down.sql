-- Drop the primary key
ALTER TABLE sites_locations DROP PRIMARY KEY;

-- Recreate the original unique key on (site, location)
ALTER TABLE sites_locations ADD UNIQUE KEY site (site, location);