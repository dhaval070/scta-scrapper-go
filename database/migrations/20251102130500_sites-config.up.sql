-- Create sites_config table for dynamic site configuration
CREATE TABLE IF NOT EXISTS sites_config (
    id INT PRIMARY KEY AUTO_INCREMENT,
    site_name VARCHAR(100) UNIQUE NOT NULL COMMENT 'Unique site identifier',
    display_name VARCHAR(200) COMMENT 'Human-readable site name',
    base_url VARCHAR(500) NOT NULL COMMENT 'Base URL for the site',
    home_team VARCHAR(100) COMMENT 'Home team name for the site',
    
    -- Parsing strategy
    parser_type ENUM('day_details', 'day_details_parser1', 'day_details_parser2', 'month_based', 'group_based', 'custom', 'external') NOT NULL COMMENT 'Type of parser to use',
    
    -- Parser configuration stored as JSON
    parser_config JSON COMMENT 'Parser-specific configuration in JSON format',
    
    -- Metadata
    enabled BOOLEAN DEFAULT true COMMENT 'Whether site scraping is enabled',
    last_scraped_at TIMESTAMP NULL COMMENT 'Last successful scrape timestamp',
    scrape_frequency_hours INT DEFAULT 24 COMMENT 'Minimum hours between scrapes',
    notes TEXT COMMENT 'Additional notes about the site',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_enabled (enabled),
    INDEX idx_parser_type (parser_type),
    INDEX idx_last_scraped (last_scraped_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Configuration for dynamically scraped sites';

-- Insert initial configurations for existing sites

-- Day Details parser sites (15 sites using ParseDayDetailsSchedule)
INSERT INTO sites_config (site_name, display_name, base_url, home_team, parser_type, parser_config) VALUES
('ckgha', 'CKGHA', 'https://ckgha.com/', 'ckgha', 'day_details', 
 JSON_OBJECT('tournament_check_exact', true, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('bghc', 'BGHC', 'https://bghc.ca/', 'bghc', 'day_details',
 JSON_OBJECT('tournament_check_exact', true, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('cygha', 'CYGHA', 'https://cygha.com/', 'cygha', 'day_details',
 JSON_OBJECT('tournament_check_exact', true, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('georginagirlshockey', 'Georgina Girls Hockey', 'https://georginagirlshockey.com/', 'georginagirlshockey', 'day_details',
 JSON_OBJECT('tournament_check_exact', true, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('lakeshorelightning', 'Lakeshore Lightning', 'https://lakeshorelightning.com/', 'lakeshorelightning', 'day_details',
 JSON_OBJECT('tournament_check_exact', true, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('londondevilettes', 'London Devilettes', 'https://londondevilettes.ca/', 'londondevilettes', 'day_details',
 JSON_OBJECT('tournament_check_exact', true, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('pgha', 'PGHA', 'https://pgha.net/', 'pgha', 'day_details',
 JSON_OBJECT('tournament_check_exact', true, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('sarniagirlshockey', 'Sarnia Girls Hockey', 'https://sarniagirlshockey.com/', 'sarniagirlshockey', 'day_details',
 JSON_OBJECT('tournament_check_exact', true, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('scarboroughsharks', 'Scarborough Sharks', 'https://scarboroughsharks.com/', 'scarboroughsharks', 'day_details',
 JSON_OBJECT('tournament_check_exact', true, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('smgha', 'SMGHA', 'https://smgha.com/', 'smgha', 'day_details',
 JSON_OBJECT('tournament_check_exact', true, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('wgha', 'WGHA', 'https://wgha.org/', 'wgha', 'day_details',
 JSON_OBJECT('tournament_check_exact', true, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('londonjuniorknights', 'London Junior Knights', 'https://londonjuniorknights.com/', 'londonjuniorknights', 'day_details',
 JSON_OBJECT('tournament_check_exact', false, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('londonjuniormustangs', 'London Junior Mustangs', 'https://londonjuniormustangs.ca/', 'londonjuniormustangs', 'day_details',
 JSON_OBJECT('tournament_check_exact', false, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('aceshockey', 'Aces Hockey', 'https://aceshockey.com/', 'aceshockey', 'day_details',
 JSON_OBJECT('tournament_check_exact', true, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d')),

('waterlooravens', 'Waterloo Ravens', 'https://waterlooravens.com/', 'waterlooravens', 'day_details',
 JSON_OBJECT('tournament_check_exact', false, 'log_errors', true, 'url_template', 'Calendar/?Month=%d&Year=%d'));

-- Group-based parser sites (22 sites using ParseSiteListGroups)
INSERT INTO sites_config (site_name, display_name, base_url, home_team, parser_type, parser_config) VALUES
('beechey', 'Beechey', 'https://beechey.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('eomhl', 'EOMHL', 'https://eomhl.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('essexll', 'Essex LL', 'https://essexll.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('fourcountieshockey', 'Four Counties Hockey', 'https://fourcountieshockey.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('gbmhl', 'GBMHL', 'https://gbmhl.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('gbtll', 'GBTLL', 'https://gbtll.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('grandriverll', 'Grand River LL', 'https://grandriverll.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('haldimandll', 'Haldimand LL', 'https://haldimandll.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('intertownll', 'Intertown LL', 'https://intertownll.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('leohockey', 'LEO Hockey', 'https://leohockey.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('lmll', 'LMLL', 'https://lmll.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('ndll', 'NDLL', 'https://ndll.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('omha-aaa', 'OMHA AAA', 'https://omha-aaa.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('srll', 'SRLL', 'https://srll.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('threecountyhockey', 'Three County Hockey', 'https://threecountyhockey.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('victoriadurham', 'Victoria Durham', 'https://victoriadurham.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('woaa.on', 'WOAA', 'https://woaa.on.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('bluewaterhockey', 'Bluewater Hockey', 'https://bluewaterhockey.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/div/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('lakeshorehockey', 'Lakeshore Hockey', 'https://lakeshorehockey.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/div/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('niagrahockey', 'Niagara Hockey', 'https://niagrahockey.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/div/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('shamrockhockey', 'Shamrock Hockey', 'https://shamrockhockey.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/div/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),

('ysmhl', 'YSMHL', 'https://ysmhl.ca/', '', 'group_based',
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/div/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')
);

-- Additional sites discovered after initial migration

-- Month-based parser sites (4 sites) - These use ParseMonthBasedSchedule
INSERT INTO sites_config (site_name, display_name, base_url, home_team, parser_type, parser_config) VALUES
('heoaaaleague', 'HEOAA League', 'https://heoaaaleague.ca/', 'heoaaaleague', 'month_based', 
 JSON_OBJECT('url_template', 'Schedule/?Month=%d&Year=%d', 'team_parse_strategy', 'subject-owner-first')),
('spfhahockey', 'SPFHA Hockey', 'https://spfhahockey.com/', 'sun parlour', 'month_based', 
 JSON_OBJECT('url_template', 'Schedule/?Month=%d&Year=%d', 'team_parse_strategy', 'first-char-detect')),
('windsoraaazone', 'Windsor AAA Zone', 'https://windsoraaazone.net/', 'Windsor AAA Zone', 'month_based', 
 JSON_OBJECT('url_template', 'Schedule/?Month=%d&Year=%d', 'team_parse_strategy', 'first-char-detect')),
('wmha', 'WMHA', 'https://wmha.net/', 'Windsor Spitfires', 'month_based', 
 JSON_OBJECT('url_template', 'Schedule/?Month=%d&Year=%d', 'team_parse_strategy', 'first-char-detect'));

-- Group-based parser sites (4 sites) - These use ParseSiteListGroups
INSERT INTO sites_config (site_name, display_name, base_url, home_team, parser_type, parser_config) VALUES
('mpshl', 'MPSHL', 'https://mpshl.ca/', '', 'group_based', 
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),
('southerncounties', 'Southern Counties', 'https://southerncounties.ca/', '', 'group_based', 
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),
('tcmhl', 'TCMHL', 'https://tcmhl.ca/', '', 'group_based', 
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/')),
('ucmhl', 'UCMHL', 'https://ucmhl.ca/', '', 'group_based', 
 JSON_OBJECT('group_xpath', '//div[@class="site-list"]/div/a', 'group_url_template', 'Groups/%s/Calendar/?Month=%d&Year=%d', 'seasons_url', 'Seasons/Current/'));


-- Parser1 sites (22 sites) - Use parser1.ParseSchedules for away game handling
INSERT INTO sites_config (site_name, display_name, base_url, home_team, parser_type, parser_config) VALUES
('arnpriorminorhockey', 'Arnprior Minor Hockey', 'https://arnpriorminorhockey.ca/', 'arnpriorminorhockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('auroraminorhockey', 'Aurora Minor Hockey', 'https://auroraminorhockey.com/', 'auroraminorhockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('barrysbayminorhockey', 'Barrys Bay Minor Hockey', 'https://barrysbayminorhockey.ca/', 'barrysbayminorhockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('burlingtoneagles', 'Burlington Eagles', 'https://burlingtoneagles.com/', 'burlingtoneagles', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('dramha', 'DRAMHA', 'https://dramha.com/', 'dramha', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('durhamcrusaders', 'Durham Crusaders', 'https://durhamcrusaders.ca/', 'durhamcrusaders', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('glha', 'GLHA', 'https://glha.ca/', 'glha', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('kitchenerminorhockey', 'Kitchener Minor Hockey', 'https://kitchenerminorhockey.com/', 'kitchenerminorhockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('londonbanditshockey', 'London Bandits Hockey', 'https://londonbanditshockey.com/', 'londonbanditshockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('northlondonhockey', 'North London Hockey', 'https://northlondonhockey.ca/', 'northlondonhockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('oakridgeaeroshockey', 'Oak Ridge Aeros Hockey', 'https://oakridgeaeroshockey.ca/', 'oakridgeaeroshockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('pembrokeminorhockey', 'Pembroke Minor Hockey', 'https://pembrokeminorhockey.com/', 'pembrokeminorhockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('petawawaminorhockey', 'Petawawa Minor Hockey', 'https://petawawaminorhockey.ca/', 'petawawaminorhockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('peterboroughhockey', 'Peterborough Hockey', 'https://peterboroughhockey.com/', 'peterboroughhockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('peterboroughminorpetes', 'Peterborough Minor Petes', 'https://peterboroughminorpetes.ca/', 'peterboroughminorpetes', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('renfrewminorhockey', 'Renfrew Minor Hockey', 'https://renfrewminorhockey.ca/', 'renfrewminorhockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('sdhockey', 'SD Hockey', 'https://sdhockey.ca/', 'sdahockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('upperottawavalleyaces', 'Upper Ottawa Valley Aces', 'https://upperottawavalleyaces.com/', 'upperottawavalleyaces', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('waldenminorhockey', 'Walden Minor Hockey', 'https://waldenminorhockey.ca/', 'waldenminorhockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('waxers', 'Waxers', 'https://waxers.com/', 'waxers', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('westlondonhockey', 'West London Hockey', 'https://westlondonhockey.ca/', 'westlondonhockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('whitbyhockey', 'Whitby Hockey', 'https://whitbyhockey.com/', 'whitbyhockey', 'day_details_parser1', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d'));


-- Parser2 sites (5 sites) - Use parser2.ParseSchedules with regex-based home/away detection
-- These sites require explicit "home game" or "away game" markers in the schedule
INSERT INTO sites_config (site_name, display_name, base_url, home_team, parser_type, parser_config) VALUES
('manitoulinminorhockey', 'Manitoulin Minor Hockey', 'https://manitoulinminorhockey.ca/', 'manitoulinminorhockey', 'day_details_parser2', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('manitoulinpanthers', 'Manitoulin Panthers', 'https://manitoulinpanthers.com/', 'manitoulinpanthers', 'day_details_parser2', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('saultmajorhockey', 'Sault Major Hockey', 'https://saultmajorhockey.ca/', 'saultmajorhockey', 'day_details_parser2', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('soopeewee', 'Soo Peewee', 'https://soopeewee.ca/', 'soopeewee', 'day_details_parser2', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('timminsminorhockey', 'Timmins Minor Hockey', 'https://timminsminorhockey.com/', 'timminsminorhockey', 'day_details_parser2', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d')),
('powassanhawks', 'Powassan Hawks', 'https://powassanhawks.com/', 'powassanhawks', 'day_details_parser2', JSON_OBJECT('url_template', 'Calendar/?Month=%d&Year=%d'));


-- External parser sites (2 sites) - Uses standalone binary
-- These sites have custom parsing logic that doesn't fit standard parsers
INSERT INTO sites_config (site_name, display_name, base_url, home_team, parser_type, parser_config, notes) VALUES
('alliancehockey', 'Alliance Hockey', 'https://alliancehockey.com/', 'alliancehockey', 'external', JSON_OBJECT('binary_path', './bin/alliancehockey'), 'Custom date parsing - calls standalone binary'),
('lugsports', 'LUG Sports', 'https://www.lugsports.com/stats#/1709/schedule?season_id=8683', 'lugsports', 'external', JSON_OBJECT('binary_path', './bin/lugsports'), 'Requires Selenium WebDriver - calls standalone binary');

-- Custom/complex sites (disabled - need manual configuration)
-- These sites have custom implementations that need individual analysis before enabling
INSERT INTO sites_config (site_name, display_name, base_url, home_team, parser_type, parser_config, enabled, notes) VALUES
('uovmhl', 'UOVMHL', 'https://uovmhl.ca/', 'uovmhl', 'custom', 
 '{}', false, 'Custom implementation - needs analysis');

-- Remaining sites (30 sites) - Disabled until verified
-- These sites exist but need individual analysis to determine correct parser type and configuration
-- After analysis, update their parser_type, parser_config and set enabled=true
INSERT INTO sites_config (site_name, display_name, base_url, home_team, parser_type, parser_config, enabled, notes) VALUES
('edinahockeyassociation', 'Edina Hockey Association', 'https://www.edinahockeyassociation.com/schedule/day/league_instance/216128/', 'edinahockeyassociation', 'custom', '{}', false, 'Needs parser type verification'),
('rockeymountainhockey', 'Rockey Mountain Hockey', '', 'rockeymountainhockey', 'custom', '{}', false, 'Needs parser type verification');
