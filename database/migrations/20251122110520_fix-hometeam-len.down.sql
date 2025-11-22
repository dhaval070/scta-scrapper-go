-- we won't reduce length back to 64
alter table events modify home_team varchar(128) not null, modify guest_team varchar(128) not null;
