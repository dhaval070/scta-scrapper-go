drop table if exists gamesheet_schedules;

delete from sites_config where site_name like '%_gs';

delete from sites where site like '%_gs';
delete from sites_locations  where site like '%_gs';
