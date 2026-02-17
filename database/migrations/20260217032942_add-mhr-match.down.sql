alter table mhr_locations drop postal_code;
alter table sites_locations drop mhr_location_id, drop mhr_match_type;
alter table events drop mhr_location_id;

-- alter table sites_locations drop index idx_mhr;
-- alter table events drop index idx_mhr;
