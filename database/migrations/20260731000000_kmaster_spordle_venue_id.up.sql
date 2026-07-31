ALTER TABLE kmaster_venue_list
  ADD COLUMN spordle_venue_id VARCHAR(36) DEFAULT NULL AFTER mhr_venue_id,
  ADD KEY idx_kmaster_spordle_venue_id (spordle_venue_id);
