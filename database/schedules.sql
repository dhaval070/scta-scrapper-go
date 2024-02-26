-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: mariadb:3306
-- Generation Time: Feb 25, 2024 at 07:38 PM
-- Server version: 11.0.3-MariaDB
-- PHP Version: 8.2.11

SET FOREIGN_KEY_CHECKS=0;
SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";

-- --------------------------------------------------------

--
-- Table structure for table `events`
--

CREATE TABLE IF NOT EXISTS `events` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `site` varchar(64) NOT NULL,
  `source_type` varchar(64) DEFAULT NULL,
  `datetime` datetime NOT NULL,
  `home_team` varchar(64) NOT NULL,
  `oid_home` varchar(128) DEFAULT NULL,
  `guest_team` varchar(64) NOT NULL,
  `oid_guest` varchar(128) DEFAULT NULL,
  `location` varchar(64) NOT NULL,
  `division` varchar(64) NOT NULL,
  `surface_id` int(11) NOT NULL DEFAULT 0,
  `date_created` timestamp NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- --------------------------------------------------------

--
-- Table structure for table `feed_modes`
--

CREATE TABLE IF NOT EXISTS `feed_modes` (
  `id` int(11) NOT NULL,
  `feed_mode` varchar(64) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `locations`
--

CREATE TABLE IF NOT EXISTS `locations` (
  `id` int(11) NOT NULL,
  `address1` text DEFAULT NULL,
  `address2` text NOT NULL,
  `city` varchar(32) DEFAULT NULL,
  `name` varchar(64) DEFAULT NULL,
  `uuid` varchar(128) DEFAULT NULL,
  `recording_hours_local` varchar(32) NOT NULL,
  `postal_code` varchar(32) NOT NULL,
  `all_sheets_count` int(11) NOT NULL DEFAULT 0,
  `longitude` float NOT NULL,
  `latitude` float NOT NULL,
  `logo_url` text NOT NULL,
  `province_id` int(11) NOT NULL,
  `venue_status` varchar(11) NOT NULL,
  `zone` varchar(32) NOT NULL,
  `total_surfaces` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `province_id` (`province_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `ohf_teams`
--

CREATE TABLE IF NOT EXISTS `ohf_teams` (
  `team_number` text DEFAULT NULL,
  `team_name` text DEFAULT NULL,
  `team_organization` text DEFAULT NULL,
  `team_organization_path` text DEFAULT NULL,
  `team_gender_identity` text DEFAULT NULL,
  `division_name` text DEFAULT NULL,
  `registrations_class_name` text DEFAULT NULL,
  `category_name` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `provinces`
--

CREATE TABLE IF NOT EXISTS `provinces` (
  `id` int(11) NOT NULL,
  `province_name` varchar(32) NOT NULL,
  `country` varchar(32) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `renditions`
--

CREATE TABLE IF NOT EXISTS `renditions` (
  `id` int(11) NOT NULL,
  `surface_id` int(11) NOT NULL,
  `name` varchar(32) NOT NULL,
  `width` int(11) NOT NULL,
  `height` int(11) NOT NULL,
  `ratio` varchar(16) NOT NULL,
  `bitrate` bigint(20) UNSIGNED NOT NULL,
  PRIMARY KEY (`id`),
  KEY `surface_id` (`surface_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `sites`
--

CREATE TABLE IF NOT EXISTS `sites` (
  `site` varchar(64) NOT NULL,
  `url` varchar(128) NOT NULL,
  PRIMARY KEY (`site`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `sites_locations`
--

CREATE TABLE IF NOT EXISTS `sites_locations` (
  `site` varchar(64) NOT NULL,
  `location` varchar(64) NOT NULL,
  `location_id` int(11) DEFAULT 0,
  `loc` varchar(64) DEFAULT NULL,
  `surface` varchar(64) DEFAULT NULL,
  `address` varchar(128) DEFAULT NULL,
  `match_type` varchar(32) DEFAULT NULL,
  `surface_id` int(11) NOT NULL DEFAULT 0,
  UNIQUE KEY `site` (`site`,`location`),
  KEY `location_id` (`location_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `surfaces`
--

CREATE TABLE IF NOT EXISTS `surfaces` (
  `id` int(11) NOT NULL,
  `location_id` int(11) NOT NULL,
  `name` varchar(64) NOT NULL,
  `uuid` varchar(128) NOT NULL,
  `orderIndex` int(11) NOT NULL,
  `venue_id` int(11) NOT NULL,
  `closed_from` bigint(20) UNSIGNED NOT NULL,
  `coming_soon` tinyint(1) NOT NULL,
  `online` tinyint(1) NOT NULL,
  `status` varchar(32) NOT NULL,
  `sports` varchar(32) NOT NULL,
  `first_media_date` bigint(20) UNSIGNED NOT NULL,
  PRIMARY KEY (`id`),
  KEY `location_id` (`location_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `surface_feed_modes`
--

CREATE TABLE IF NOT EXISTS `surface_feed_modes` (
  `surface_id` int(11) NOT NULL,
  `feed_mode_id` int(11) NOT NULL,
  PRIMARY KEY (`surface_id`,`feed_mode_id`),
  KEY `surface_id` (`surface_id`),
  KEY `feed_mode_id` (`feed_mode_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `todb_surfaces`
--

CREATE TABLE IF NOT EXISTS `todb_surfaces` (
  `id` int(11) NOT NULL,
  `fullname` varchar(256) DEFAULT NULL,
  `fullshortname` varchar(256) DEFAULT NULL,
  `street` varchar(256) DEFAULT NULL,
  `city` varchar(256) DEFAULT NULL,
  `province` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_unicode_ci;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `locations`
--
ALTER TABLE `locations`
  ADD CONSTRAINT `locations_ibfk_2` FOREIGN KEY (`province_id`) REFERENCES `provinces` (`id`);

--
-- Constraints for table `renditions`
--
ALTER TABLE `renditions`
  ADD CONSTRAINT `renditions_ibfk_1` FOREIGN KEY (`surface_id`) REFERENCES `surfaces` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `sites_locations`
--
ALTER TABLE `sites_locations`
  ADD CONSTRAINT `sites_locations_ibfk_1` FOREIGN KEY (`site`) REFERENCES `sites` (`site`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `surfaces`
--
ALTER TABLE `surfaces`
  ADD CONSTRAINT `surfaces_ibfk_1` FOREIGN KEY (`location_id`) REFERENCES `locations` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `surface_feed_modes`
--
ALTER TABLE `surface_feed_modes`
  ADD CONSTRAINT `surface_feed_modes_ibfk_1` FOREIGN KEY (`surface_id`) REFERENCES `surfaces` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `surface_feed_modes_ibfk_2` FOREIGN KEY (`feed_mode_id`) REFERENCES `feed_modes` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;
SET FOREIGN_KEY_CHECKS=1;
COMMIT;
