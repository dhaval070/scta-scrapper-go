package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"calendar-scrapper/config"
	"calendar-scrapper/pkg/cmdutil"
	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/scraper"
	"calendar-scrapper/pkg/siteconfig"

	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	siteNames       string
	allSites        bool
	dueOnly         bool
	workers         int
	dateFlag        string
	outfile         string
	importLocations bool
)

var rootCmd = &cobra.Command{
	Use:   "scraper",
	Short: "Calendar scraper for hockey scheduling sites",
	Long:  `A CLI tool to scrape hockey schedule data from various scheduling sites and store it in a database.`,
	Run:   runScraper,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured sites",
	Long:  `List all sites configured in the database with their status and last scraped time.`,
	Run: func(cmd *cobra.Command, args []string) {
		listConfiguredSites()
	},
}

func init() {
	rootCmd.Flags().StringVar(&siteNames, "sites", "", "Comma-separated list of site names to scrape")
	rootCmd.Flags().BoolVar(&allSites, "all", false, "Scrape all enabled sites")
	rootCmd.Flags().BoolVar(&dueOnly, "due", false, "Only scrape sites due for scraping (based on frequency)")
	rootCmd.Flags().IntVar(&workers, "workers", 1, "Number of sites to scrape in parallel (default: 1)")
	rootCmd.Flags().StringVar(&dateFlag, "date", "", "Date in MMYYYY or MM/YYYY format (default: current month)")
	rootCmd.Flags().StringVar(&outfile, "outfile", "", "Output file path for scraped data")
	rootCmd.Flags().BoolVar(&importLocations, "import-locations", false, "Import locations to database")

	rootCmd.AddCommand(listCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runScraper(cmd *cobra.Command, args []string) {
	// Validate input
	if siteNames == "" && !allSites && !dueOnly {
		log.Fatal("Error: Must specify --sites=<name1,name2,...>, --all, or --due flag")
	}

	if siteNames != "" && (allSites || dueOnly) {
		log.Fatal("Error: Cannot use --sites with --all or --due")
	}

	// Validate workers count
	if workers < 1 {
		log.Fatal("Error: workers must be at least 1")
	}
	if workers > 20 {
		log.Fatal("Error: workers cannot exceed 20")
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

	if allSites {
		sites, err = loader.GetAllEnabled()
		if err != nil {
			log.Fatalf("Failed to load enabled sites: %v", err)
		}
		log.Printf("Loaded %d enabled sites\n", len(sites))
	} else if dueOnly {
		sites, err = loader.GetDueForScraping()
		if err != nil {
			log.Fatalf("Failed to load sites due for scraping: %v", err)
		}
		log.Printf("Loaded %d sites due for scraping\n", len(sites))
	} else {
		// Parse comma-separated site names
		names := strings.Split(siteNames, ",")
		for i, name := range names {
			names[i] = strings.TrimSpace(name)
		}
		log.Printf("Loading %d specified site(s): %v\n", len(names), names)
		for _, name := range names {
			site, err := loader.GetSite(name)
			if err != nil {
				log.Fatalf("Failed to load site '%s': %v", name, err)
			}
			sites = append(sites, *site)
		}
	}

	if len(sites) == 0 {
		log.Println("No sites to process")
		return
	}

	// Determine date range
	mm, yyyy := getMonthYear(dateFlag)

	log.Printf("Using %d worker(s) for parallel scraping\n", workers)

	// Create flags struct for compatibility with existing code
	flags := &cmdutil.Flags{
		Date:            &dateFlag,
		Outfile:         &outfile,
		ImportLocations: &importLocations,
	}

	// Process sites using worker pool
	successCount, failCount := processSitesWithPool(sites, loader, mm, yyyy, flags, workers)

	// Summary
	log.Printf("\n========================================")
	log.Printf("SUMMARY")
	log.Printf("========================================")
	log.Printf("Total sites: %d", len(sites))
	log.Printf("Successful: %d", successCount)
	log.Printf("Failed: %d", failCount)
	log.Printf("========================================\n")
}

// processSitesWithPool processes multiple sites using a worker pool
func processSitesWithPool(sites []siteconfig.SiteConfig, loader *siteconfig.Loader, mm, yyyy int, flags *cmdutil.Flags, workers int) (int, int) {
	// Create channels
	jobs := make(chan siteconfig.SiteConfig, len(sites))
	type result struct {
		siteName string
		success  bool
		err      error
	}
	results := make(chan result, len(sites))

	// Start workers
	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for site := range jobs {
				log.Printf("\n[Worker %d] ========================================", workerID)
				log.Printf("[Worker %d] Processing: %s (%s)", workerID, site.DisplayName, site.SiteName)
				log.Printf("[Worker %d] ========================================", workerID)

				err := processSite(&site, loader, mm, yyyy, flags)
				if err != nil {
					log.Printf("[Worker %d] ❌ ERROR scraping %s: %v\n", workerID, site.SiteName, err)
					results <- result{siteName: site.SiteName, success: false, err: err}
					continue
				}

				// Update last scraped timestamp
				if err := loader.UpdateLastScraped(site.ID); err != nil {
					log.Printf("[Worker %d] ⚠️  Warning: Failed to update last_scraped_at for %s: %v\n", workerID, site.SiteName, err)
				}

				log.Printf("[Worker %d] ✅ Successfully processed %s\n", workerID, site.SiteName)
				results <- result{siteName: site.SiteName, success: true, err: nil}
			}
		}(w)
	}

	// Send jobs to workers
	for _, site := range sites {
		jobs <- site
	}
	close(jobs)

	// Wait for all workers to complete
	wg.Wait()
	close(results)

	// Collect results
	successCount := 0
	failCount := 0
	for res := range results {
		if res.success {
			successCount++
		} else {
			failCount++
		}
	}

	return successCount, failCount
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
		dir := filepath.Dir(*flags.Outfile)
		basename := filepath.Base(*flags.Outfile)

		outfile := filepath.Join(dir, fmt.Sprintf("%s_%s", site.SiteName, basename))
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
