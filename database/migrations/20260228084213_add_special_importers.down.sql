-- Remove special importers added in up migration
DELETE FROM sites_config WHERE site_name IN ('gthl', 'nyhl', 'mhl');