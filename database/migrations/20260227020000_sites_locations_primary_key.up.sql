-- Drop the existing unique key on (site, location)
ALTER TABLE sites_locations DROP INDEX site;

-- Add primary key on (site, location)
ALTER TABLE sites_locations ADD PRIMARY KEY (site, location);