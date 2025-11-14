create table gamesheet_schedules(
    id bigint unsigned auto_increment primary key,
    season_id int unsigned not null,
    game_data json not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp on update current_timestamp,
    key idx_season_id (season_id)
)engine=innodb charset=utf8;
