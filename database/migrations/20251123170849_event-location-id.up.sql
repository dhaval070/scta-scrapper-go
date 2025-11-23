alter table events add location_id int default 0 after division;

update events e, surfaces s set e.location_id=s.location_id where e.surface_id=s.id;
