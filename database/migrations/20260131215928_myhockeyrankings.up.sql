SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";

CREATE TABLE `mhr_locations` (
  `mhr_id` int NOT NULL,
  `rink_name` varchar(128) NOT NULL,
  `aka` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `address` varchar(256) NOT NULL,
  `phone` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `website` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `streaming` varchar(128) DEFAULT NULL,
  `notes` text,
  `livebarn_installed` tinyint NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

ALTER TABLE `mhr_locations`
  ADD PRIMARY KEY (`mhr_id`);
COMMIT;


SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";

CREATE TABLE `mhr_sheet` (
  `id` int NOT NULL,
  `rink_name` varchar(128) NOT NULL,
  `livebarn_installed` tinyint DEFAULT '0',
  `liebarn_venue_id` int DEFAULT '0',
  `rink_type` varchar(128) DEFAULT NULL,
  `alt_name` varchar(128) DEFAULT NULL,
  `alt_name2` varchar(128) DEFAULT NULL,
  `alt_name3` varchar(128) DEFAULT NULL,
  `rink_pads` int DEFAULT '0',
  `city` varchar(128) NOT NULL,
  `state` varchar(128) NOT NULL,
  `address` varchar(128) DEFAULT NULL,
  `zip` varchar(64) DEFAULT NULL,
  `country` varchar(128) DEFAULT NULL,
  `phone` varchar(32) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

ALTER TABLE `mhr_sheet`
  ADD PRIMARY KEY (`id`);
COMMIT;
