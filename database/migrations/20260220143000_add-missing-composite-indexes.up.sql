-- Add composite indexes to optimize site-specific filtering queries in repository
-- These indexes target the WHERE site=? AND column=? patterns used extensively in pkg/repository/repo.go

-- Optimize queries: WHERE site=? AND location_id=0 (line 236) and WHERE site=? AND location_id>0 (line 279)
CREATE INDEX idx_sites_locations_site_location_id ON sites_locations(site, location_id) ALGORITHM=INPLACE;

-- Optimize queries: WHERE site=? AND mhr_location_id=0 (line 730) and similar MHR matching patterns
CREATE INDEX idx_sites_locations_site_mhr_location_id ON sites_locations(site, mhr_location_id) ALGORITHM=INPLACE;

-- Optimize queries: WHERE site=? AND surface_id=0 AND location_id>0 (line 289, 300) and similar surface matching
CREATE INDEX idx_sites_locations_site_surface_id_location_id ON sites_locations(site, surface_id, location_id) ALGORITHM=INPLACE;