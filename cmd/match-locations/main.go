package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/repository"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var cfg config.Config
var repo *repository.Repository

var (
	siteNames string
	allSites  bool
)

var rootCmd = &cobra.Command{
	Use:   "match-locations",
	Short: "Match sites locations with livebarn or mhr locations",
	Run:   runMatcher,
}

func init() {
	config.Init("config", ".")
	cfg = config.MustReadConfig()

	repo = repository.NewRepository(cfg)

	rootCmd.Flags().StringVar(&siteNames, "sites", "", "Comma-separated list of site names to scrape")
	rootCmd.Flags().BoolVar(&allSites, "all", false, "Scrape all enabled sites")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runMatcher(cmd *cobra.Command, args []string) {
	if siteNames == "" && !allSites {
		log.Fatal("Error: Must specify --sites=<name1,name2,...>, --all, or --due flag")
	}

	if siteNames != "" && allSites {
		log.Fatal("Error: Cannot use --sites with --all or --due")
	}

	var names []string

	if allSites {
		err := repo.DB.Raw(`select site_name FROM sites_config`).Scan(&names).Error
		if err != nil {
			log.Fatal(err)
		}
	} else {
		names = strings.Split(siteNames, ",")
		for i, name := range names {
			names[i] = strings.TrimSpace(name)
		}
	}

	log.Printf("Processing %d specified site(s): %v\n", len(names), names)

	for _, name := range names {
		r := repo.Site(name)

		err := r.ImportLoc([]model.SitesLocation{})
		if err != nil {
			log.Println(err)
		}
	}
}
