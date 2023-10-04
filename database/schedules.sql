-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: mariadb:3306
-- Generation Time: Oct 04, 2023 at 09:05 PM
-- Server version: 11.0.3-MariaDB
-- PHP Version: 8.2.11

SET FOREIGN_KEY_CHECKS=0;
SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `schedules`
--

-- --------------------------------------------------------

--
-- Table structure for table `feed_modes`
--

DROP TABLE IF EXISTS `feed_modes`;
CREATE TABLE `feed_modes` (
  `id` int(11) NOT NULL,
  `feed_mode` varchar(64) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `locations`
--

DROP TABLE IF EXISTS `locations`;
CREATE TABLE `locations` (
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
  `zone` varchar(32) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `provinces`
--

DROP TABLE IF EXISTS `provinces`;
CREATE TABLE `provinces` (
  `id` int(11) NOT NULL,
  `province_name` varchar(32) NOT NULL,
  `country` varchar(32) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `renditions`
--

DROP TABLE IF EXISTS `renditions`;
CREATE TABLE `renditions` (
  `id` int(11) NOT NULL,
  `surface_id` int(11) NOT NULL,
  `name` varchar(32) NOT NULL,
  `width` int(11) NOT NULL,
  `height` int(11) NOT NULL,
  `ratio` varchar(16) NOT NULL,
  `bitrate` bigint(10) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `sites`
--

DROP TABLE IF EXISTS `sites`;
CREATE TABLE `sites` (
  `site` varchar(64) NOT NULL,
  `url` varchar(128) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `sites`
--

INSERT INTO `sites` (`site`, `url`) VALUES
('alliancehockey', 'https://alliancehockey.com/'),
('omha-aaa', 'https://omha-aaa.net/Seasons/Current/');

-- --------------------------------------------------------

--
-- Table structure for table `sites_locations`
--

DROP TABLE IF EXISTS `sites_locations`;
CREATE TABLE `sites_locations` (
  `site` varchar(64) NOT NULL,
  `location` varchar(64) NOT NULL,
  `location_id` int(11) DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `surfaces`
--

DROP TABLE IF EXISTS `surfaces`;
CREATE TABLE `surfaces` (
  `id` int(11) NOT NULL,
  `location_id` int(11) NOT NULL,
  `name` varchar(64) NOT NULL,
  `uuid` varchar(128) NOT NULL,
  `orderIndex` int(11) NOT NULL,
  `venue_id` int(11) NOT NULL,
  `closed_from` bigint(10) UNSIGNED NOT NULL,
  `coming_soon` tinyint(1) NOT NULL,
  `online` tinyint(1) NOT NULL,
  `status` varchar(32) NOT NULL,
  `sports` varchar(32) NOT NULL,
  `first_media_date` bigint(10) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `surface_feed_modes`
--

DROP TABLE IF EXISTS `surface_feed_modes`;
CREATE TABLE `surface_feed_modes` (
  `surface_id` int(11) NOT NULL,
  `feed_mode_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `feed_modes`
--
ALTER TABLE `feed_modes`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `locations`
--
ALTER TABLE `locations`
  ADD PRIMARY KEY (`id`),
  ADD KEY `province_id` (`province_id`);

--
-- Indexes for table `provinces`
--
ALTER TABLE `provinces`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `renditions`
--
ALTER TABLE `renditions`
  ADD PRIMARY KEY (`id`),
  ADD KEY `surface_id` (`surface_id`);

--
-- Indexes for table `sites`
--
ALTER TABLE `sites`
  ADD PRIMARY KEY (`site`);

--
-- Indexes for table `sites_locations`
--
ALTER TABLE `sites_locations`
  ADD UNIQUE KEY `site` (`site`,`location`),
  ADD KEY `location_id` (`location_id`);

--
-- Indexes for table `surfaces`
--
ALTER TABLE `surfaces`
  ADD PRIMARY KEY (`id`),
  ADD KEY `location_id` (`location_id`);

--
-- Indexes for table `surface_feed_modes`
--
ALTER TABLE `surface_feed_modes`
  ADD PRIMARY KEY (`surface_id`,`feed_mode_id`),
  ADD KEY `surface_id` (`surface_id`),
  ADD KEY `feed_mode_id` (`feed_mode_id`);

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

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
