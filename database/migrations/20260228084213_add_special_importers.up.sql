-- Add special importers (gthl, nyhl, mhl) as external parsers
-- These sites use league-specific mapping tables and output CSV in 7-column format

INSERT INTO sites_config (site_name, display_name, base_url, home_team, parser_type, parser_config, notes) VALUES
('gthl', 'GTHL', 'https://example.com/', 'gthl', 'external', JSON_OBJECT('binary_path', './bin/gthl-import'), 'Special importer using internal API, outputs CSV with empty address column'),
('nyhl', 'NYHL', 'https://example.com/', 'nyhl', 'external', JSON_OBJECT('binary_path', './bin/nyhl-import'), 'Special importer using internal API, outputs CSV with empty address column'),
('mhl', 'MHL', 'https://example.com/', 'mhl', 'external', JSON_OBJECT('binary_path', './bin/mhl-import'), 'Special importer using internal API, outputs CSV with empty address column')
ON DUPLICATE KEY UPDATE
    display_name = VALUES(display_name),
    base_url = VALUES(base_url),
    home_team = VALUES(home_team),
    parser_type = VALUES(parser_type),
    parser_config = VALUES(parser_config),
    notes = VALUES(notes);