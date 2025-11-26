-- phpMyAdmin SQL Dump
-- version 5.2.3
-- https://www.phpmyadmin.net/
--
-- Host: db
-- Generation Time: Nov 26, 2025 at 10:30 AM
-- Server version: 8.0.43
-- PHP Version: 8.3.26

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";

--
-- Database: `schedules`
--

--
-- Dumping data for table `gamesheet_seasons`
--

INSERT INTO `gamesheet_seasons` (`id`, `title`, `site`, `league_id`, `is_active`, `start_date`, `end_date`) VALUES
(10239, '2025-26 Tier II Regular Season', 'gs_202526TIRS', NULL, 1, NULL, NULL),
(10294, 'Delaware Valley Hockey League - 2025-2026', 'gs_DVHL20252026', NULL, 1, NULL, NULL),
(10301, '2025-26 AAHA - Exhibition/Scrimmage Season', 'gs_202526AESS', NULL, 1, NULL, NULL),
(10302, '2025-26 AZYHL Regular Season', 'gs_202526ARS', NULL, 1, NULL, NULL),
(10312, 'Maryland Student Hockey League - Regular Season 25/26', 'gs_MSHLRS2526', NULL, 1, NULL, NULL),
(10358, 'Eastern Junior Elite Prospects League 2025-2026', 'gs_EJEPL20252026', NULL, 1, NULL, NULL),
(10397, 'SESCL - Southeastern Showcase League - 2025/2026', 'gs_SSSL20252026', NULL, 1, NULL, NULL),
(10399, 'SYTHL - Southern Youth Travel Hockey League - 2025/2026', 'gs_SSYTHL20252026', NULL, 1, NULL, NULL),
(10450, 'Inter County Scholastic Hockey League - 2025', 'gs_ICSHL2025', NULL, 1, NULL, NULL),
(10482, 'Carolinas Hockey League - CHL - 2025/2026', 'gs_CHLC20252026', NULL, 1, NULL, NULL),
(10548, 'Alaska State Hockey Association - Regular Season 2025/2026', 'gs_ASHARS20252026', NULL, 1, NULL, NULL),
(10570, 'Mid-Atlantic Hockey Alliance - AA Season 2025-2026', 'gs_MAHAAS20252026', NULL, 1, NULL, NULL),
(10604, 'Suburban Adult Hockey League Macomb: Fall/Winter 2025-2026', 'gs_SAHLMFW20252026', NULL, 1, NULL, NULL),
(10605, 'Suburban Adult Hockey League Rochester: Fall/Winter 2025-2026', 'gs_SAHLRFW20252026', NULL, 1, NULL, NULL),
(10612, 'Michigan Amateur Youth Hockey League - Exhibition Games 2025/2026', 'gs_MAYHLEG20252026', NULL, 1, NULL, NULL),
(10634, 'MAHA - Exhibition Games - 2025-2026', 'gs_MEG20252026', NULL, 1, NULL, NULL),
(10658, 'Michigan Girls Hockey League - Season 25/26', 'gs_MGHLS2526', NULL, 1, NULL, NULL),
(10667, 'New York State Amateur Hockey Association (NYSAHA) - \"Q\"/TB Games Season 2025/2026', 'gs_NYSAHANQTGS20252026', NULL, 1, NULL, NULL),
(10668, 'New York State Amateur Hockey Association (NYSAHA) - 2025/2026', 'gs_NYSAHAN20252026', NULL, 1, NULL, NULL),
(10697, 'Wisconsin Amateur Hockey Association - Regular Season 2025/2026', 'gs_WAHARS20252026', NULL, 1, NULL, NULL),
(10740, 'Metro Girls High School Hockey League 2025/2026', 'gs_MGHSHL20252026', NULL, 1, NULL, NULL),
(10754, 'Vermont State Amateur Hockey Association 2025-2026', 'gs_VSAHA20252026', NULL, 1, NULL, NULL),
(10757, 'Atlantic Coast Hockey Conference - 2025-2026', 'gs_ACHC20252026', NULL, 1, NULL, NULL),
(10761, 'Des Moines Youth Hockey Association - 2025/2026', 'gs_DMYHA20252026', NULL, 1, NULL, NULL),
(10774, 'Royal Oak: Fall/Winter Regular Season - 2025-2026', 'gs_ROFWRS20252026', NULL, 1, NULL, NULL),
(10804, 'BioSteel Sports Academy 2025-2026 Season', 'gs_BSA20252026S', NULL, 1, NULL, NULL),
(10809, 'Independent Teams, Tournament Teams, & Exhibition Games 2025-2026', 'gs_ITTTEG20252026', NULL, 1, NULL, NULL),
(10844, 'Mullett Hockey League - Fall/Winter 2025-2026', 'gs_MHLFW20252026', NULL, 1, NULL, NULL),
(10857, 'Minnesota Hockey - Districts Exhibition Season 2025-2026', 'gs_MHDES20252026', NULL, 1, NULL, NULL),
(11105, 'Greater New York Stars 2025-2026 - To Be Deleted', 'gs_GNYS20252026TBD', NULL, 1, NULL, NULL),
(11139, 'Hampton Roads House League Fall 2025', 'gs_HRHLF2025', NULL, 1, NULL, NULL),
(11177, 'Ohio AAA Blue Jackets - Regular Season 15, 16U, 18U - 2025/26', 'gs_OABJRS1511202526', NULL, 1, NULL, NULL),
(11193, '2025/26 - Regular Season - AAU College Hockey', 'gs_202526RSACH', NULL, 1, NULL, NULL),
(11223, 'Bay County Adult Hockey League - Fall 2025', 'gs_BCAHLF2025', NULL, 1, NULL, NULL),
(11225, 'WAHL Winter League - 2025-2026', 'gs_WWL20252026', NULL, 1, NULL, NULL),
(11246, 'Minnesota District 4 - 2025-2026', 'gs_MD420252026', NULL, 1, NULL, NULL),
(11248, 'Minnesota - District 12 - 2025/2026', 'gs_MD1220252026', NULL, 1, NULL, NULL),
(11252, 'Indiana State High School Hockey Association - 2025/2026 Season', 'gs_ISHSHA20252026S', NULL, 1, NULL, NULL),
(11255, 'Jamestown Amateur Hockey League - 2025-2026 Season', 'gs_JAHL20252026S', NULL, 1, NULL, NULL),
(11281, 'Jr. Coyotes - 2025-2026', 'gs_JC20252026', NULL, 1, NULL, NULL),
(11282, 'CDP Scottsdale - 2025-2026', 'gs_CS20252026', NULL, 1, NULL, NULL),
(11283, 'CDP Chandler - 2025-2026', 'gs_CC20252026', NULL, 1, NULL, NULL),
(11420, 'Michigan High School Hockey League - 2025/2026 Regular Season', 'gs_MHSHL20252026RS', NULL, 1, NULL, NULL),
(11437, 'Capital Corridor Hockey League - 2025-2026', 'gs_CCHL20252026', NULL, 1, NULL, NULL),
(11440, 'Buffalo Vibes Draft League - Fall/Winter 2025-2026', 'gs_BVDLFW20252026', NULL, 1, NULL, NULL),
(11491, '404 Interlock Houseleague 2025-2026', 'gs_404IH20252026', NULL, 1, NULL, NULL),
(11550, 'Prowl Hockey - Regular Season 2025/26', 'gs_PHRS202526', NULL, 1, NULL, NULL),
(11577, 'Gateway High School Hockey League - Regular Season 2025/2026', 'gs_GHSHLRS20252026', NULL, 1, NULL, NULL),
(11579, 'Hudson Valley High School Hockey League 2025-2026', 'gs_HVHSHL20252026', NULL, 1, NULL, NULL),
(11619, 'Michigan Amateur Youth Hockey League - Regular Season 2025/2026', 'gs_MAYHLRS20252026', NULL, 1, NULL, NULL),
(11628, 'Westman High School Hockey League - 2025-2026', 'gs_WHSHL20252026', NULL, 1, NULL, NULL),
(11629, 'Buckeye Travel Hockey League - Regular Season 2025/2026', 'gs_BTHLRS20252026', NULL, 1, NULL, NULL),
(11666, 'Flin Flon Minor Hockey - Season 2025-2026', 'gs_FFMHS20252026', NULL, 1, NULL, NULL),
(11708, 'Hockey Tonk (AAA & AA) - Nov 7-10, 2025', 'gs_HTAAN7102025', NULL, 1, NULL, NULL),
(11719, 'Ice House Adult League - Fall/Winter 2025', 'gs_IHALFW2025', NULL, 1, NULL, NULL),

(11769, 'Charleston Youth Hockey - Regular Season 2025-2026', 'gs_CYHRS20252026', NULL, 1, NULL, NULL),
(11790, 'AACPS MS Hockey - 2025-2026', 'gs_AMH20252026', NULL, 1, NULL, NULL),
(11837, 'Granite State League - 2025-2026 Playoff Season', 'gs_GSL20252026PS', NULL, 1, NULL, NULL),
(11881, 'St. Cloud - Municipal Athletic Complex - 2025/2026', 'gs_SCMAC20252026', NULL, 1, NULL, NULL),
(11899, 'Richland Youth Hockey - 2025/2026 Season', 'gs_RYH20252026S', NULL, 1, NULL, NULL),
(11902, 'CUP Hockey League 2025/2026', 'gs_CHL20252026', NULL, 1, NULL, NULL),
(11930, 'Steamboat - Adele Mountain Divas - Week 1 - Nov 7-9, 2025', 'gs_SAMDW1N792025', NULL, 1, NULL, NULL),
(11931, 'Steamboat - Adele Mountain Divas - Week 2 - Nov 14-16, 2025', 'gs_SAMDW2N14162025', NULL, 1, NULL, NULL),
(11939, 'West Shore - Fenstermachers Cup 10UB - Nov 7-9, 2025', 'gs_WSFC1N792025', NULL, 1, NULL, NULL),
(11965, 'Nassau County Hockey 2025-2026', 'gs_NCH20252026', NULL, 1, NULL, NULL),
(11977, 'Adirondack Youth Hockey - Fire on Ice - Nov 7-9, 2025', 'gs_AYHFOIN792025', NULL, 1, NULL, NULL),
(11981, 'West Coast Elite Showcase - Nov 7-9, 2025', 'gs_WCESN792025', NULL, 1, NULL, NULL),
(11982, 'D9 Albert Lea - B Bantam Tournament - Nov 7-9, 2025', 'gs_DALBBTN792025', NULL, 1, NULL, NULL),
(11994, 'D9 Albert Lea - A Bantam Tournament - Nov 14-16, 2025', 'gs_DALABTN14162025', NULL, 1, NULL, NULL),
(12023, 'NED - 15O Tier 1 Regionals - 2025/26', 'gs_Ned15OT1R202526', NULL, 1, NULL, NULL),
(12024, 'NED - 16U Tier 1 Regionals - 2025/26', 'gs_Ned16UT1R202526', NULL, 1, NULL, NULL),
(12025, 'NED - 18U Tier 1 Regionals - 2025/26', 'gs_Ned18UT1R202526', NULL, 1, NULL, NULL),
(12093, '53nd Annual Carman Cougars Hockey Tournament - Nov 14-15, 2025', 'gs_5ACCHTN14152025', NULL, 1, NULL, NULL),
(12261, 'Big Rapids - 10UB Tournament - Nov 14-16, 2025', 'gs_BR1TN14162025', NULL, 1, NULL, NULL),
(12278, 'Stowe Youth Hockey - Stick Season Showdown -Nov 13-16, 2025', 'gs_SYHSSSN13162025', NULL, 1, NULL, NULL),
(12319, 'Sun Prairie - Squirt B / PeeWee B Groundhog Games Tournament - Nov 15-16, 2025', 'gs_SPSBPBGGTN15162025', NULL, 1, NULL, NULL),
(12330, 'Butte - Jake Siddoway Memorial Bantam Tournament - Nov 14-16, 2025', 'gs_BJSMBTN14162025', NULL, 1, NULL, NULL);
COMMIT;

-- phpMyAdmin SQL Dump
-- version 5.2.3
-- https://www.phpmyadmin.net/
--
-- Host: db
-- Generation Time: Nov 26, 2025 at 10:31 AM
-- Server version: 8.0.43
-- PHP Version: 8.3.26

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";

--
-- Database: `schedules`
--

--
-- Dumping data for table `sites_config`
--

INSERT INTO `sites_config` (`site_name`, `display_name`, `base_url`, `home_team`, `parser_type`, `parser_config`, `enabled`, `last_scraped_at`, `scrape_frequency_hours`, `notes`, `created_at`, `updated_at`) VALUES
('gs_202526TIRS', '2025-26 Tier II Regular Season', '', NULL, 'external', '{\"season_id\": 10239, \"extra_args\": [\"--sites=gs_202526TIRS \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_DVHL20252026', 'Delaware Valley Hockey League - 2025-2026', '', NULL, 'external', '{\"season_id\": 10294, \"extra_args\": [\"--sites=gs_DVHL20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_202526AESS', '2025-26 AAHA - Exhibition/Scrimmage Season', '', NULL, 'external', '{\"season_id\": 10301, \"extra_args\": [\"--sites=gs_202526AESS \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_202526ARS', '2025-26 AZYHL Regular Season', '', NULL, 'external', '{\"season_id\": 10302, \"extra_args\": [\"--sites=gs_202526ARS \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_MSHLRS2526', 'Maryland Student Hockey League - Regular Season 25/26', '', NULL, 'external', '{\"season_id\": 10312, \"extra_args\": [\"--sites=gs_MSHLRS2526 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_EJEPL20252026', 'Eastern Junior Elite Prospects League 2025-2026', '', NULL, 'external', '{\"season_id\": 10358, \"extra_args\": [\"--sites=gs_EJEPL20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_SSSL20252026', 'SESCL - Southeastern Showcase League - 2025/2026', '', NULL, 'external', '{\"season_id\": 10397, \"extra_args\": [\"--sites=gs_SSSL20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_SSYTHL20252026', 'SYTHL - Southern Youth Travel Hockey League - 2025/2026', '', NULL, 'external', '{\"season_id\": 10399, \"extra_args\": [\"--sites=gs_SSYTHL20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_ICSHL2025', 'Inter County Scholastic Hockey League - 2025', '', NULL, 'external', '{\"season_id\": 10450, \"extra_args\": [\"--sites=gs_ICSHL2025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_CHLC20252026', 'Carolinas Hockey League - CHL - 2025/2026', '', NULL, 'external', '{\"season_id\": 10482, \"extra_args\": [\"--sites=gs_CHLC20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_ASHARS20252026', 'Alaska State Hockey Association - Regular Season 2025/2026', '', NULL, 'external', '{\"season_id\": 10548, \"extra_args\": [\"--sites=gs_ASHARS20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_MAHAAS20252026', 'Mid-Atlantic Hockey Alliance - AA Season 2025-2026', '', NULL, 'external', '{\"season_id\": 10570, \"extra_args\": [\"--sites=gs_MAHAAS20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_SAHLMFW20252026', 'Suburban Adult Hockey League Macomb: Fall/Winter 2025-2026', '', NULL, 'external', '{\"season_id\": 10604, \"extra_args\": [\"--sites=gs_SAHLMFW20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_SAHLRFW20252026', 'Suburban Adult Hockey League Rochester: Fall/Winter 2025-2026', '', NULL, 'external', '{\"season_id\": 10605, \"extra_args\": [\"--sites=gs_SAHLRFW20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_MAYHLEG20252026', 'Michigan Amateur Youth Hockey League - Exhibition Games 2025/2026', '', NULL, 'external', '{\"season_id\": 10612, \"extra_args\": [\"--sites=gs_MAYHLEG20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_MEG20252026', 'MAHA - Exhibition Games - 2025-2026', '', NULL, 'external', '{\"season_id\": 10634, \"extra_args\": [\"--sites=gs_MEG20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_MGHLS2526', 'Michigan Girls Hockey League - Season 25/26', '', NULL, 'external', '{\"season_id\": 10658, \"extra_args\": [\"--sites=gs_MGHLS2526 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_NYSAHANQTGS20252026', 'New York State Amateur Hockey Association (NYSAHA) - \"Q\"/TB Games Season 2025/2026', '', NULL, 'external', '{\"season_id\": 10667, \"extra_args\": [\"--sites=gs_NYSAHANQTGS20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_NYSAHAN20252026', 'New York State Amateur Hockey Association (NYSAHA) - 2025/2026', '', NULL, 'external', '{\"season_id\": 10668, \"extra_args\": [\"--sites=gs_NYSAHAN20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_WAHARS20252026', 'Wisconsin Amateur Hockey Association - Regular Season 2025/2026', '', NULL, 'external', '{\"season_id\": 10697, \"extra_args\": [\"--sites=gs_WAHARS20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_MGHSHL20252026', 'Metro Girls High School Hockey League 2025/2026', '', NULL, 'external', '{\"season_id\": 10740, \"extra_args\": [\"--sites=gs_MGHSHL20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_VSAHA20252026', 'Vermont State Amateur Hockey Association 2025-2026', '', NULL, 'external', '{\"season_id\": 10754, \"extra_args\": [\"--sites=gs_VSAHA20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_ACHC20252026', 'Atlantic Coast Hockey Conference - 2025-2026', '', NULL, 'external', '{\"season_id\": 10757, \"extra_args\": [\"--sites=gs_ACHC20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_DMYHA20252026', 'Des Moines Youth Hockey Association - 2025/2026', '', NULL, 'external', '{\"season_id\": 10761, \"extra_args\": [\"--sites=gs_DMYHA20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_ROFWRS20252026', 'Royal Oak: Fall/Winter Regular Season - 2025-2026', '', NULL, 'external', '{\"season_id\": 10774, \"extra_args\": [\"--sites=gs_ROFWRS20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_BSA20252026S', 'BioSteel Sports Academy 2025-2026 Season', '', NULL, 'external', '{\"season_id\": 10804, \"extra_args\": [\"--sites=gs_BSA20252026S \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_ITTTEG20252026', 'Independent Teams, Tournament Teams, & Exhibition Games 2025-2026', '', NULL, 'external', '{\"season_id\": 10809, \"extra_args\": [\"--sites=gs_ITTTEG20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_MHLFW20252026', 'Mullett Hockey League - Fall/Winter 2025-2026', '', NULL, 'external', '{\"season_id\": 10844, \"extra_args\": [\"--sites=gs_MHLFW20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_MHDES20252026', 'Minnesota Hockey - Districts Exhibition Season 2025-2026', '', NULL, 'external', '{\"season_id\": 10857, \"extra_args\": [\"--sites=gs_MHDES20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_GNYS20252026TBD', 'Greater New York Stars 2025-2026 - To Be Deleted', '', NULL, 'external', '{\"season_id\": 11105, \"extra_args\": [\"--sites=gs_GNYS20252026TBD \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_HRHLF2025', 'Hampton Roads House League Fall 2025', '', NULL, 'external', '{\"season_id\": 11139, \"extra_args\": [\"--sites=gs_HRHLF2025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_OABJRS1511202526', 'Ohio AAA Blue Jackets - Regular Season 15, 16U, 18U - 2025/26', '', NULL, 'external', '{\"season_id\": 11177, \"extra_args\": [\"--sites=gs_OABJRS1511202526 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_202526RSACH', '2025/26 - Regular Season - AAU College Hockey', '', NULL, 'external', '{\"season_id\": 11193, \"extra_args\": [\"--sites=gs_202526RSACH \"], \"binary_path\": \"./bin/gamesheet\"}', 1, '2025-11-26 10:08:34', 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:34'),
('gs_BCAHLF2025', 'Bay County Adult Hockey League - Fall 2025', '', NULL, 'external', '{\"season_id\": 11223, \"extra_args\": [\"--sites=gs_BCAHLF2025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_WWL20252026', 'WAHL Winter League - 2025-2026', '', NULL, 'external', '{\"season_id\": 11225, \"extra_args\": [\"--sites=gs_WWL20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_MD420252026', 'Minnesota District 4 - 2025-2026', '', NULL, 'external', '{\"season_id\": 11246, \"extra_args\": [\"--sites=gs_MD420252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_MD1220252026', 'Minnesota - District 12 - 2025/2026', '', NULL, 'external', '{\"season_id\": 11248, \"extra_args\": [\"--sites=gs_MD1220252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_ISHSHA20252026S', 'Indiana State High School Hockey Association - 2025/2026 Season', '', NULL, 'external', '{\"season_id\": 11252, \"extra_args\": [\"--sites=gs_ISHSHA20252026S \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_JAHL20252026S', 'Jamestown Amateur Hockey League - 2025-2026 Season', '', NULL, 'external', '{\"season_id\": 11255, \"extra_args\": [\"--sites=gs_JAHL20252026S \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_JC20252026', 'Jr. Coyotes - 2025-2026', '', NULL, 'external', '{\"season_id\": 11281, \"extra_args\": [\"--sites=gs_JC20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_CS20252026', 'CDP Scottsdale - 2025-2026', '', NULL, 'external', '{\"season_id\": 11282, \"extra_args\": [\"--sites=gs_CS20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_CC20252026', 'CDP Chandler - 2025-2026', '', NULL, 'external', '{\"season_id\": 11283, \"extra_args\": [\"--sites=gs_CC20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_MHSHL20252026RS', 'Michigan High School Hockey League - 2025/2026 Regular Season', '', NULL, 'external', '{\"season_id\": 11420, \"extra_args\": [\"--sites=gs_MHSHL20252026RS \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_CCHL20252026', 'Capital Corridor Hockey League - 2025-2026', '', NULL, 'external', '{\"season_id\": 11437, \"extra_args\": [\"--sites=gs_CCHL20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_BVDLFW20252026', 'Buffalo Vibes Draft League - Fall/Winter 2025-2026', '', NULL, 'external', '{\"season_id\": 11440, \"extra_args\": [\"--sites=gs_BVDLFW20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_404IH20252026', '404 Interlock Houseleague 2025-2026', '', NULL, 'external', '{\"season_id\": 11491, \"extra_args\": [\"--sites=gs_404IH20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_PHRS202526', 'Prowl Hockey - Regular Season 2025/26', '', NULL, 'external', '{\"season_id\": 11550, \"extra_args\": [\"--sites=gs_PHRS202526 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_GHSHLRS20252026', 'Gateway High School Hockey League - Regular Season 2025/2026', '', NULL, 'external', '{\"season_id\": 11577, \"extra_args\": [\"--sites=gs_GHSHLRS20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_HVHSHL20252026', 'Hudson Valley High School Hockey League 2025-2026', '', NULL, 'external', '{\"season_id\": 11579, \"extra_args\": [\"--sites=gs_HVHSHL20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_MAYHLRS20252026', 'Michigan Amateur Youth Hockey League - Regular Season 2025/2026', '', NULL, 'external', '{\"season_id\": 11619, \"extra_args\": [\"--sites=gs_MAYHLRS20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_WHSHL20252026', 'Westman High School Hockey League - 2025-2026', '', NULL, 'external', '{\"season_id\": 11628, \"extra_args\": [\"--sites=gs_WHSHL20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_BTHLRS20252026', 'Buckeye Travel Hockey League - Regular Season 2025/2026', '', NULL, 'external', '{\"season_id\": 11629, \"extra_args\": [\"--sites=gs_BTHLRS20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_FFMHS20252026', 'Flin Flon Minor Hockey - Season 2025-2026', '', NULL, 'external', '{\"season_id\": 11666, \"extra_args\": [\"--sites=gs_FFMHS20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_HTAAN7102025', 'Hockey Tonk (AAA & AA) - Nov 7-10, 2025', '', NULL, 'external', '{\"season_id\": 11708, \"extra_args\": [\"--sites=gs_HTAAN7102025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_IHALFW2025', 'Ice House Adult League - Fall/Winter 2025', '', NULL, 'external', '{\"season_id\": 11719, \"extra_args\": [\"--sites=gs_IHALFW2025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),

('gs_CYHRS20252026', 'Charleston Youth Hockey - Regular Season 2025-2026', '', NULL, 'external', '{\"season_id\": 11769, \"extra_args\": [\"--sites=gs_CYHRS20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_AMH20252026', 'AACPS MS Hockey - 2025-2026', '', NULL, 'external', '{\"season_id\": 11790, \"extra_args\": [\"--sites=gs_AMH20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_GSL20252026PS', 'Granite State League - 2025-2026 Playoff Season', '', NULL, 'external', '{\"season_id\": 11837, \"extra_args\": [\"--sites=gs_GSL20252026PS \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_SCMAC20252026', 'St. Cloud - Municipal Athletic Complex - 2025/2026', '', NULL, 'external', '{\"season_id\": 11881, \"extra_args\": [\"--sites=gs_SCMAC20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_RYH20252026S', 'Richland Youth Hockey - 2025/2026 Season', '', NULL, 'external', '{\"season_id\": 11899, \"extra_args\": [\"--sites=gs_RYH20252026S \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_CHL20252026', 'CUP Hockey League 2025/2026', '', NULL, 'external', '{\"season_id\": 11902, \"extra_args\": [\"--sites=gs_CHL20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_SAMDW1N792025', 'Steamboat - Adele Mountain Divas - Week 1 - Nov 7-9, 2025', '', NULL, 'external', '{\"season_id\": 11930, \"extra_args\": [\"--sites=gs_SAMDW1N792025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_SAMDW2N14162025', 'Steamboat - Adele Mountain Divas - Week 2 - Nov 14-16, 2025', '', NULL, 'external', '{\"season_id\": 11931, \"extra_args\": [\"--sites=gs_SAMDW2N14162025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_WSFC1N792025', 'West Shore - Fenstermachers Cup 10UB - Nov 7-9, 2025', '', NULL, 'external', '{\"season_id\": 11939, \"extra_args\": [\"--sites=gs_WSFC1N792025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_NCH20252026', 'Nassau County Hockey 2025-2026', '', NULL, 'external', '{\"season_id\": 11965, \"extra_args\": [\"--sites=gs_NCH20252026 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_AYHFOIN792025', 'Adirondack Youth Hockey - Fire on Ice - Nov 7-9, 2025', '', NULL, 'external', '{\"season_id\": 11977, \"extra_args\": [\"--sites=gs_AYHFOIN792025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_WCESN792025', 'West Coast Elite Showcase - Nov 7-9, 2025', '', NULL, 'external', '{\"season_id\": 11981, \"extra_args\": [\"--sites=gs_WCESN792025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_DALBBTN792025', 'D9 Albert Lea - B Bantam Tournament - Nov 7-9, 2025', '', NULL, 'external', '{\"season_id\": 11982, \"extra_args\": [\"--sites=gs_DALBBTN792025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_DALABTN14162025', 'D9 Albert Lea - A Bantam Tournament - Nov 14-16, 2025', '', NULL, 'external', '{\"season_id\": 11994, \"extra_args\": [\"--sites=gs_DALABTN14162025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_Ned15OT1R202526', 'NED - 15O Tier 1 Regionals - 2025/26', '', NULL, 'external', '{\"season_id\": 12023, \"extra_args\": [\"--sites=gs_Ned15OT1R202526 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_Ned16UT1R202526', 'NED - 16U Tier 1 Regionals - 2025/26', '', NULL, 'external', '{\"season_id\": 12024, \"extra_args\": [\"--sites=gs_Ned16UT1R202526 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_Ned18UT1R202526', 'NED - 18U Tier 1 Regionals - 2025/26', '', NULL, 'external', '{\"season_id\": 12025, \"extra_args\": [\"--sites=gs_Ned18UT1R202526 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_5ACCHTN14152025', '53nd Annual Carman Cougars Hockey Tournament - Nov 14-15, 2025', '', NULL, 'external', '{\"season_id\": 12093, \"extra_args\": [\"--sites=gs_5ACCHTN14152025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_BR1TN14162025', 'Big Rapids - 10UB Tournament - Nov 14-16, 2025', '', NULL, 'external', '{\"season_id\": 12261, \"extra_args\": [\"--sites=gs_BR1TN14162025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_SYHSSSN13162025', 'Stowe Youth Hockey - Stick Season Showdown -Nov 13-16, 2025', '', NULL, 'external', '{\"season_id\": 12278, \"extra_args\": [\"--sites=gs_SYHSSSN13162025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_SPSBPBGGTN15162025', 'Sun Prairie - Squirt B / PeeWee B Groundhog Games Tournament - Nov 15-16, 2025', '', NULL, 'external', '{\"season_id\": 12319, \"extra_args\": [\"--sites=gs_SPSBPBGGTN15162025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10'),
('gs_BJSMBTN14162025', 'Butte - Jake Siddoway Memorial Bantam Tournament - Nov 14-16, 2025', '', NULL, 'external', '{\"season_id\": 12330, \"extra_args\": [\"--sites=gs_BJSMBTN14162025 \"], \"binary_path\": \"./bin/gamesheet\"}', 1, NULL, 24, NULL, '2025-11-26 10:08:10', '2025-11-26 10:08:10');
COMMIT;
