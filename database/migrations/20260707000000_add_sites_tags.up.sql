CREATE TABLE IF NOT EXISTS `sites_tags` (
  `site_name`  varchar(100) NOT NULL,
  `tag_id`     int(11)      NOT NULL,
  `created_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`site_name`, `tag_id`),
  KEY `idx_sites_tags_tag_id` (`tag_id`),
  CONSTRAINT `fk_sites_tags_site_name` FOREIGN KEY (`site_name`) REFERENCES `sites_config` (`site_name`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_sites_tags_tag_id` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
