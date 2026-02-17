alter table mhr_locations add postal_code varchar(16) not null;
alter table sites_locations add mhr_location_id int not null default 0, add mhr_match_type varchar(32) null;
alter table events add mhr_location_id int not null;

alter table sites_locations add index idx_mhr(mhr_location_id);
alter table events add index idx_mhr(mhr_location_id);
