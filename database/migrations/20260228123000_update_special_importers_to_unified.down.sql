-- Revert special importers back to separate binaries
UPDATE sites_config 
SET parser_config = JSON_OBJECT('binary_path', './bin/gthl-import')
WHERE site_name = 'gthl' AND parser_type = 'external';

UPDATE sites_config 
SET parser_config = JSON_OBJECT('binary_path', './bin/nyhl-import')
WHERE site_name = 'nyhl' AND parser_type = 'external';

UPDATE sites_config 
SET parser_config = JSON_OBJECT('binary_path', './bin/mhl-import')
WHERE site_name = 'mhl' AND parser_type = 'external';