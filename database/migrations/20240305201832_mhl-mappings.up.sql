SET FOREIGN_KEY_CHECKS=0;
SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


--
-- Table structure for table `mhl_mappings`
--

CREATE TABLE `mhl_mappings` (
  `location` varchar(64) NOT NULL,
  `surface_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `mhl_mappings`
--

INSERT INTO `mhl_mappings` (`location`, `surface_id`) VALUES
('Amesbury', 3876),
('Angela James Arena', 4067),
('Angus Glen East', 4323),
('Angus Glen West', 4324),
('Baycrest', 3859),
('Bayview', 3858),
('Brampton Memorial Arena', 4065),
('Cassie Cambell Comm Cent', 4348),
('Chesswood 1', 3734),
('Chesswood 2', 3735),
('Chesswood 3', 3736),
('Chesswood 4', 3737),
('Chic Murray', 4394),
('Clarkson', 4220),
('Commander Park 1', 4091),
('Commander Park 2', 4092),
('Cummer Park', 4135),
('Don Mills Arena', 4113),
('Don Montgomery 1', 4133),
('Don Montgomery 2', 4134),
('Downsview', 4195),
('Erin Mills 1', 3834),
('Erin Mills 2', 3835),
('Etobicoke Centennial East', 3874),
('Etobicoke Centennial West ', 3875),
('Etobicoke Ice Sports 1', 2293),
('Etobicoke Ice Sports 2', 2294),
('Etobicoke Ice Sports 3', 2295),
('Etobicoke Ice Sports 4', 2296),
('Fenside', 4142),
('Ford Performance Centre 1', 4124),
('Ford Performance Centre 2', 4125),
('Ford Performance Centre 3', 4126),
('Ford Performance Centre 4', 4127),
('Forest Hill/Larry Grossman', 4437),
('Glen Abbey Blue', 4110),
('Glen Abbey Green', 4111),
('Gord and Irene Risk', 3878),
('Goulding Park', 3879),
('Grandravine', 3880),
('Habitant', 3857),
('Herbert Carnegie', 4196),
('Huron Park', 4266),
('Huron Park Arena', 4266),
('Iceland 1', 4214),
('Iceland 2', 4215),
('Iceland 3', 4216),
('Iceland 4', 4217),
('Iroquios Park 1', 4474),
('Iroquios Park 3', 4476),
('Iroquois Park 4', 4477),
('Lambton', 4198),
('Leaside Memorial A', 3705),
('Leaside Memorial B', 3754),
('Malvern 1', 4071),
('Malvern 2', 4070),
('Maple Arena', 4136),
('MasterCard Centre 1', 4124),
('MasterCard Centre 2', 4125),
('MasterCard Centre 3', 4126),
('MasterCard Centre 4', 4127),
('McKinney 1', 4321),
('Meadowvale 1', 4578),
('Meadowvale 2', 4579),
('Meadowvale 3', 4580),
('Meadowvale 4', 4581),
('North Toronto', 4590),
('Oakville Ice Sports 1', 2289),
('Oakville Ice Sports 2', 2290),
('Oakville Ice Sports 3', 2291),
('Oakville Ice Sports 4', 2292),
('Oriole', 3881),
('Paramount', 4210),
('Paramount Fine Foods 2', 4211),
('Paramount Fine Foods 3', 4212),
('Paramount Fine Foods 4', 4213),
('Paramount Fine Foods Centre 2', 4211),
('Paramount Fine Foods Centre 3', 4212),
('Paramount Fine Foods Centre 4', 4213),
('Paramount Fine Foods Centre Main Bowl', 4210),
('Paramount Fine Foods Main Bowl', 4210),
('Pinepoint', 4202),
('Pleasantview', 4114),
('Port Credit', 4385),
('Port Credit Arena', 4385),
('Powerade - Tire World', 131),
('Scarborough Centennial Arena', 4197),
('Scarborough Ice Sports 1', 2314),
('Scarborough Ice Sports 2', 2315),
('Scarborough Ice Sports 3', 2316),
('Scarborough Ice Sports 4', 2317),
('Scotiabank Pond 1', 4087),
('Scotiabank Pond 2', 4088),
('Scotiabank Pond 3', 4089),
('Scotiabank Pond 4', 4090),
('St. Michaels', 4386),
('Thornhill Centre West', 4424),
('Tomken 1', 4452),
('Tomken 2', 4453),
('Trafalgar Park', 4115),
('Trisan Centre', 4222),
('Valleys', 4219),
('Vaughan Sports Village A', 4120),
('Vaughan Sports Village B', 4121),
('Vaughan Sports Village C', 4122),
('Vaughan Sports Village D', 4123),
('Vic Johnston', 4066),
('Victoria Village', 4426),
('Westwood 1', 3552),
('Westwood 2', 3730),
('Westwood 3', 3731),
('Westwood 4', 3732),
('Westwood 5', 3733),
('York 1', 2297),
('York 2', 2298),
('York 3', 2299),
('York 4', 2300),
('York 5', 2301),
('York 6', 2302);

--
-- Indexes for dumped tables
--

--
-- Indexes for table `mhl_mappings`
--
ALTER TABLE `mhl_mappings`
  ADD PRIMARY KEY (`location`);
SET FOREIGN_KEY_CHECKS=1;
COMMIT;
