package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"calendar-scrapper/config"
	"calendar-scrapper/pkg/cmdutil"
	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/scraper"
	"calendar-scrapper/pkg/siteconfig"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Define scraper-specific flags
	siteName := flag.String("site", "", "Site name to scrape (use with --all to scrape all sites)")
	allSites := flag.Bool("all", false, "Scrape all enabled sites")
	dueOnly := flag.Bool("due", false, "Only scrape sites due for scraping (based on frequency)")
	listSites := flag.Bool("list", false, "List all configured sites and exit")

	// Parse common flags (date, outfile, import-locations)
	flags := cmdutil.ParseCommonFlags()

	// Handle list command
	if *listSites {
		listConfiguredSites()
		return
	}

	// Validate input
	if *siteName == "" && !*allSites && !*dueOnly {
		log.Fatal("Error: Must specify --site=<name>, --all, or --due flag")
	}

	if *siteName != "" && (*allSites || *dueOnly) {
		log.Fatal("Error: Cannot use --site with --all or --due")
	}

	// Initialize configuration
	config.Init("config", ".")
	cfg := config.MustReadConfig()

	// Connect to database
	db, err := initDB(&cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	loader := siteconfig.NewLoader(db)

	// Get sites to process
	var sites []siteconfig.SiteConfig

	if *allSites {
		sites, err = loader.GetAllEnabled()
		if err != nil {
			log.Fatalf("Failed to load enabled sites: %v", err)
		}
		log.Printf("Loaded %d enabled sites\n", len(sites))
	} else if *dueOnly {
		sites, err = loader.GetDueForScraping()
		if err != nil {
			log.Fatalf("Failed to load sites due for scraping: %v", err)
		}
		log.Printf("Loaded %d sites due for scraping\n", len(sites))
	} else {
		site, err := loader.GetSite(*siteName)
		if err != nil {
			log.Fatalf("Failed to load site '%s': %v", *siteName, err)
		}
		sites = []siteconfig.SiteConfig{*site}
	}

	if len(sites) == 0 {
		log.Println("No sites to process")
		return
	}

	// Determine date range
	mm, yyyy := getMonthYear(*flags.Date)

	// Process each site
	successCount := 0
	failCount := 0

	for _, site := range sites {
		log.Printf("\n========================================")
		log.Printf("Processing: %s (%s)", site.DisplayName, site.SiteName)
		log.Printf("========================================")

		err := processSite(&site, loader, mm, yyyy, flags)
		if err != nil {
			log.Printf("❌ ERROR scraping %s: %v\n", site.SiteName, err)
			failCount++
			continue
		}

		// Update last scraped timestamp
		if err := loader.UpdateLastScraped(site.ID); err != nil {
			log.Printf("⚠️  Warning: Failed to update last_scraped_at for %s: %v\n", site.SiteName, err)
		}

		successCount++
		log.Printf("✅ Successfully processed %s\n", site.SiteName)
	}

	// Summary
	log.Printf("\n========================================")
	log.Printf("SUMMARY")
	log.Printf("========================================")
	log.Printf("Total sites: %d", len(sites))
	log.Printf("Successful: %d", successCount)
	log.Printf("Failed: %d", failCount)
	log.Printf("========================================\n")
}

// processSite handles scraping and output for a single site
func processSite(site *siteconfig.SiteConfig, loader *siteconfig.Loader, mm, yyyy int, flags *cmdutil.Flags) error {
	// Create scraper
	s, err := scraper.New(site, loader)
	if err != nil {
		return fmt.Errorf("failed to create scraper: %w", err)
	}

	// Perform scraping
	result, err := s.Scrape(mm, yyyy)
	if err != nil {
		return fmt.Errorf("scrape failed: %w", err)
	}

	if len(result) == 0 {
		log.Printf("⚠️  Warning: No schedule data found for %s\n", site.SiteName)
		return nil
	}

	log.Printf("Retrieved %d schedule entries\n", len(result))

	// Import locations if requested
	if *flags.ImportLocations {
		log.Printf("Importing locations to database...\n")
		if err := cmdutil.ImportLocations(site.SiteName, result); err != nil {
			return fmt.Errorf("failed to import locations: %w", err)
		}
		log.Printf("✓ Locations imported\n")
	}

	// Write output if requested
	if *flags.Outfile != "" {
		outfile := fmt.Sprintf("%s_%s", site.SiteName, *flags.Outfile)
		log.Printf("Writing output to: %s\n", outfile)
		if err := cmdutil.WriteOutput(outfile, result); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		log.Printf("✓ Output written\n")
	}

	return nil
}

// getMonthYear returns month and year from date flag or current date
func getMonthYear(dateFlag string) (int, int) {
	if dateFlag == "" {
		now := time.Now()
		return int(now.Month()), now.Year()
	}

	mm, yyyy := parser.ParseMonthYear(dateFlag)
	return mm, yyyy
}

// initDB initializes database connection
func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DbDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// listConfiguredSites lists all sites in the database
func listConfiguredSites() {
	config.Init("config", ".")
	cfg := config.MustReadConfig()

	db, err := initDB(&cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	siteconfig.NewLoader(db)

	var allSites []siteconfig.SiteConfig
	err = db.Order("enabled DESC, site_name ASC").Find(&allSites).Error
	if err != nil {
		log.Fatalf("Failed to load sites: %v", err)
	}

	fmt.Println("\n╔════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    CONFIGURED SITES                                ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════╝\n")

	enabledCount := 0
	disabledCount := 0

	for _, site := range allSites {
		status := "✓"
		if !site.Enabled {
			status = "✗"
			disabledCount++
		} else {
			enabledCount++
		}

		lastScraped := "Never"
		if site.LastScrapedAt != nil {
			lastScraped = site.LastScrapedAt.Format("2006-01-02 15:04")
		}

		fmt.Printf("%s %-25s %-12s %s\n",
			status,
			site.SiteName,
			site.ParserType,
			lastScraped,
		)
	}

	fmt.Printf("\n────────────────────────────────────────────────────────────────────\n")
	fmt.Printf("Total: %d sites (%d enabled, %d disabled)\n",
		len(allSites), enabledCount, disabledCount)
	fmt.Printf("────────────────────────────────────────────────────────────────────\n\n")
}
