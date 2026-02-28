-- Update special importers (gthl, nyhl, mhl) to use unified agilex binary with league flag
-- This replaces the three separate binaries with a single binary that uses --league flag

UPDATE sites_config 
SET parser_config = JSON_OBJECT(
    'binary_path', './bin/agilex',
    'extra_args', JSON_ARRAY('--league', 'gthl')
)
WHERE site_name = 'gthl' AND parser_type = 'external';

UPDATE sites_config 
SET parser_config = JSON_OBJECT(
    'binary_path', './bin/agilex',
    'extra_args', JSON_ARRAY('--league', 'nyhl')
)
WHERE site_name = 'nyhl' AND parser_type = 'external';

UPDATE sites_config 
SET parser_config = JSON_OBJECT(
    'binary_path', './bin/agilex',
    'extra_args', JSON_ARRAY('--league', 'mhl')
)
WHERE site_name = 'mhl' AND parser_type = 'external';