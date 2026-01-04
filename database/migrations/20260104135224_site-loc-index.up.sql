ALTER TABLE events ADD COLUMN edate DATE GENERATED ALWAYS AS (DATE(datetime)) STORED, ADD INDEX
   idx_events_edate_loc_site (edate, location_id, site);
