CREATE TABLE `gamesheet_leagues` (
    `id` int UNSIGNED NOT NULL,
    `league_name` varchar(128) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

ALTER TABLE `gamesheet_leagues`
    ADD PRIMARY KEY (`id`);
