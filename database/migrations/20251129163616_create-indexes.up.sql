CREATE INDEX idx_events_site_datetime ON events(site, datetime);
CREATE INDEX idx_locations_name ON locations(name);
CREATE INDEX idx_sites_locations_surface_id ON sites_locations(surface_id);
CREATE INDEX idx_provinces_name ON provinces(province_name);
