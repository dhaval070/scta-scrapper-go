-- Drop redundant single-column indexes that are superseded by composite indexes
-- After adding composite indexes (site, column) pairs, single-column indexes on those columns
-- are no longer needed for optimal query performance since all queries include site filter.

-- Drop idx_mhr: Superseded by idx_sites_locations_site_mhr_location_id (site, mhr_location_id)
-- All queries using mhr_location_id also filter by site (see pkg/repository/repo.go lines 730, 761, 774, etc.)
DROP INDEX idx_mhr ON sites_locations;

-- Drop idx_sites_locations_surface_id: Superseded by idx_sites_locations_site_surface_id_location_id (site, surface_id, location_id)
-- All queries using surface_id also filter by site (see pkg/repository/repo.go lines 289, 300, 431, etc.)
-- The composite index provides better selectivity for site-specific surface matching
DROP INDEX idx_sites_locations_surface_id ON sites_locations;

-- Note: location_id index is retained because:
-- 1. It may be used for joins without site filter (e.g., JOIN surfaces ON location_id)
-- 2. It's used in some UPDATE queries with complex WHERE clauses
-- 3. Cardinality is good (1,044 distinct values) and index size is reasonable