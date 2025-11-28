CREATE TABLE `sites` (
  `site` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `url` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`site`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE sites_locations DROP FOREIGN KEY  fk_sites_locations_site_config;

INSERT INTO sites (site, url)
SELECT
    sc.site_name,
    sc.base_url
FROM sites_config sc
LEFT JOIN sites s ON sc.site_name = s.site
WHERE s.site IS NULL;

ALTER TABLE sites_locations  ADD
   CONSTRAINT sites_locations_ibfk_1  FOREIGN KEY (site)  REFERENCES
   sites(site)  ON UPDATE CASCADE  ON DELETE RESTRICT;
