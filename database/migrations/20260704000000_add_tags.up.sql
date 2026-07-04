CREATE TABLE IF NOT EXISTS `tags` (
  `id`          int(11)      NOT NULL AUTO_INCREMENT,
  `name`        varchar(64)  NOT NULL,
  `color`       varchar(7)   DEFAULT NULL COMMENT 'Hex color code, e.g. #ff0000',
  `description` varchar(255) DEFAULT NULL,
  `created_at`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tags_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `sites_location_tags` (
  `site`       varchar(64)  NOT NULL,
  `location`   varchar(128) NOT NULL,
  `tag_id`     int(11)      NOT NULL,
  `created_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`site`, `location`, `tag_id`),
  KEY `idx_slt_tag_id` (`tag_id`),
  CONSTRAINT `fk_slt_site_location` FOREIGN KEY (`site`, `location`) REFERENCES `sites_locations` (`site`, `location`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_slt_tag_id` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
