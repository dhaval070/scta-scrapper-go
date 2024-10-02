create table users(
    username varchar(16) not null primary key,
    password varchar(64) not null,
    created_at timestamp default CURRENT_TIMESTAMP,
    updated_at timestamp default CURRENT_TIMESTAMP
);
