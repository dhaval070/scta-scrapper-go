-- phpMyAdmin SQL Dump
-- version 5.2.3
-- https://www.phpmyadmin.net/
--
-- Host: db
-- Generation Time: Nov 15, 2025 at 10:43 AM
-- Server version: 8.0.43
-- PHP Version: 8.3.26

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";

--
-- Database: `schedules`
--

-- --------------------------------------------------------

--
-- Table structure for table `gamesheet_seasons`
--

CREATE TABLE `gamesheet_seasons` (
  `id` int UNSIGNED NOT NULL,
  `title` varchar(200) DEFAULT NULL,
  `site` varchar(100) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NOT NULL,
  `league_id` int UNSIGNED DEFAULT NULL,
  `is_active` tinyint DEFAULT '0',
  `start_date` date DEFAULT NULL,
  `end_date` date DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

--
-- Dumping data for table `gamesheet_seasons`
--

INSERT INTO `gamesheet_seasons` (`id`, `title`, `site`, `league_id`, `is_active`, `start_date`, `end_date`) VALUES
(6425, 'Eastern Hockey Federation - 2024-2025', 'ehf202425_gs', 301633, 0, '2024-08-16', '2025-08-15'),
(6742, 'New England Girls Hockey League Fall 2024-2025', 'neghlf202425_gs', 314982, 0, '2024-08-16', '2025-08-15'),
(7809, 'New England Girls Hockey League Winter 2024-2025', 'neghlw202425_gs', 314982, 0, '2024-08-16', '2025-08-15'),
(9340, 'Eastern Hockey Federation Full Season Playoffs - 2024-2025', 'ehffsp202425_gs', 301633, 0, '2024-08-16', '2025-08-15'),
(9938, '2025 - 2026 Season Parity', 'sp202526_gs', 301633, 0, '2025-04-04', '2025-05-12'),
(10425, 'Premier Hockey League of New England (Season 1) 2025-2026', 'phlone202526_gs', 321869, 1, '2025-08-01', '2026-03-31'),
(10477, 'Eastern Hockey Federation - 2025-2026', 'ehf202526_gs', 301633, 1, '2025-08-01', '2026-08-15'),
(10664, 'New England Girls Hockey League Fall 2025-2026', 'neghlf202526_gs', 314982, 0, '2025-08-15', '2026-11-30'),
(10665, 'New England Girls Hockey League Winter 2025-2026', 'neghlw202526_gs', 314982, 1, '2025-12-01', '2026-03-31'),
(10783, 'OMHA AAA 2025-2026 Season', 'omha202526_gs', 1147945, 1, '2025-09-05', '2026-04-30'),
(11761, 'South Jersey High School Hockey League - 2025/2026 Season', 'sjhshl202526_gs', 1148480, 1, '2025-08-16', '2026-08-15');

--
-- Indexes for dumped tables
--

--
-- Indexes for table `gamesheet_seasons`
--
ALTER TABLE `gamesheet_seasons`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `gamesheet_seasons_site_IDX` (`site`) USING BTREE;
COMMIT;

