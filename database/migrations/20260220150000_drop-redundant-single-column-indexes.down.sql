-- Recreate single-column indexes that were dropped (for migration rollback)
-- These indexes may be less optimal than composite indexes but provide backward compatibility

-- Recreate idx_mhr index on mhr_location_id (originally created in 20260217032942_add-mhr-match.up.sql)
CREATE INDEX idx_mhr ON sites_locations(mhr_location_id);

-- Recreate idx_sites_locations_surface_id index on surface_id (originally created in 20251129163616_create-indexes.up.sql)
CREATE INDEX idx_sites_locations_surface_id ON sites_locations(surface_id);