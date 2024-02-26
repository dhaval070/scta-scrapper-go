-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: mariadb:3306
-- Generation Time: Feb 25, 2024 at 10:49 PM
-- Server version: 11.0.3-MariaDB
-- PHP Version: 8.2.11

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";

-- --------------------------------------------------------

--
-- Table structure for table `nyhl_mappings`
--

CREATE TABLE IF NOT EXISTS `nyhl_mappings` (
  `location` varchar(64) NOT NULL,
  `surface_id` int(11) NOT NULL,
  PRIMARY KEY (`location`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `nyhl_mappings`
--

INSERT INTO `nyhl_mappings` (`location`, `surface_id`) VALUES
('Amesbury', 3876),
('Chesswood 1', 3734),
('Chesswood 2', 3735),
('Commander Park R 1', 4091),
('Downsview', 4195),
('East York', 3877),
('Fenside', 4142),
('Ford Performance Centre R3', 4126),
('Forest Hill', 4437),
('Goulding Park', 3879),
('Habitant', 3857),
('Lambton', 4198),
('Leaside B', 3754),
('Pleasantview', 4114),
('Scotia Bank Pond 2', 4088),
('Vgn Sports Village B', 4121),
('Victoria Village', 4426),
('Westwood 3', 3731),
('Westwood 4', 3732);
COMMIT;
