ALTER TABLE sites_locations DROP FOREIGN KEY  sites_locations_ibfk_1;

ALTER TABLE sites_locations  ADD
   CONSTRAINT fk_sites_locations_site_config  FOREIGN KEY (site)  REFERENCES
   sites_config(site_name)  ON UPDATE CASCADE  ON DELETE RESTRICT;

drop table sites;
