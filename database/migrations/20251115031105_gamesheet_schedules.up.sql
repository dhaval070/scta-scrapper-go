create table gamesheet_schedules(
    id bigint unsigned auto_increment primary key,
    season_id int unsigned not null,
    game_data json not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp on update current_timestamp,
    key idx_season_id (season_id)
)engine=innodb charset=utf8;


insert into sites_config (site_name, display_name, base_url, parser_type, enabled, parser_config)
select site, title, '', "external", is_active, concat( '{"season_id":', id , '}') from gamesheet_seasons gs;
