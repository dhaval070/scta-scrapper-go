package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"calendar-scrapper/config"
	"calendar-scrapper/internal/schimport"
	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/repository"

	"github.com/antchfx/htmlquery"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var cmd = &cobra.Command{
	Use:   "nyhl-import",
	Short: "Import nyhl schedule",
	RunE: func(c *cobra.Command, args []string) error {
		return runNyhl()
	},
}

var (
	cfg             config.Config
	repo            *repository.Repository
	infile          *string
	sdate           *string
	dateFlag        *string
	outfile         *string
	importLocations *bool
)

func init() {
	config.Init("config", ".")
	cfg = config.MustReadConfig()
	repo = repository.NewRepository(cfg)

	infile = cmd.Flags().StringP("file", "f", "", "CSV file path (required)")
	sdate = cmd.Flags().StringP("cutoffdate", "d", "", "date-from to import events (required) . e.g. -cutoffdate 2023-01-01")

	// New flags for external parser interface
	dateFlag = cmd.Flags().String("date", "", "Month and year in mmyyyy format (e.g., 022025)")
	outfile = cmd.Flags().String("outfile", "", "Output file (use '-' for stdout)")
	importLocations = cmd.Flags().Bool("import-locations", false, "Import locations to database")
}

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

// gameToCSVRow converts a Game to 7-column CSV format
func gameToCSVRow(site string, g schimport.Game) []string {
	// Format datetime as "2006-1-02 15:04"
	// Parse the date and time
	timeStr := g.StartTime
	if strings.Contains(timeStr, "PM") {
		// Convert PM times
		parts := strings.Split(strings.TrimSuffix(timeStr, " PM"), ":")
		if len(parts) == 2 {
			hour, _ := strconv.Atoi(parts[0])
			if hour < 12 {
				hour += 12
			}
			timeStr = fmt.Sprintf("%02d:%s", hour, parts[1])
		}
	} else if strings.Contains(timeStr, "AM") {
		timeStr = strings.TrimSuffix(timeStr, " AM")
		parts := strings.Split(timeStr, ":")
		if len(parts) == 2 && parts[0] == "12" {
			timeStr = fmt.Sprintf("00:%s", parts[1])
		}
	}

	datetime := fmt.Sprintf("%s %s", g.StartDate, timeStr)
	// Convert to standard format
	t, err := time.Parse("2006-01-02 15:04", datetime)
	if err != nil {
		// Try alternative format
		t, err = time.Parse("1/2/2006 15:04", datetime)
		if err != nil {
			log.Printf("Warning: cannot parse datetime %s: %v", datetime, err)
			t = time.Now()
		}
	}

	formattedDatetime := t.Format("2006-1-02 15:04")

	return []string{
		formattedDatetime,             // datetime
		site,                          // site
		g.Home,                        // home team
		g.Visitor,                     // guest team
		g.Rink,                        // location
		g.Division + " " + g.Category, // division
		"",                            // address (empty for special importers)
	}
}

func runExternalParserMode() error {
	// Warn about import-locations flag (not supported for special importers)
	if *importLocations {
		log.Println("Warning: --import-locations not supported for special importers (gthl, nyhl, mhl)")
	}

	// Validate date flag
	if *dateFlag == "" {
		return fmt.Errorf("--date flag is required when --outfile is specified")
	}

	// Parse mmyyyy date using parser package (panics on invalid input)
	mm, yyyy := parser.ParseMonthYear(*dateFlag)

	// Use first day of month as cutoff date for API
	cdate := time.Date(yyyy, time.Month(mm), 1, 0, 0, 0, 0, time.UTC)

	// Get mappings
	m, err := repo.GetNyhlMappings()
	if err != nil {
		return err
	}

	importer := schimport.NewImporter(repo, cfg.ApiKey, cfg.ImportUrl)

	// Fetch data from API
	b, err := importer.FetchJson("nyhl", cdate)
	if err != nil {
		return err
	}

	var data schimport.Data
	if err = json.Unmarshal(b, &data); err != nil {
		return err
	}

	if len(data.Games) == 0 {
		log.Println("no games to import")
		return nil
	}

	// Prepare CSV output
	var output io.Writer
	if *outfile == "-" {
		output = os.Stdout
	} else {
		f, err := os.Create(*outfile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		output = f
	}

	writer := csv.NewWriter(output)

	// Process games
	for _, game := range data.Games {
		sid, ok := m[game.Rink]
		if !ok || sid == 0 {
			log.Printf("skipping unmapped location: %s", game.Rink)
			continue
		}

		row := gameToCSVRow("nyhl", game)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("CSV flush error: %w", err)
	}

	return nil
}

func runNyhl() error {
	// External parser mode: output CSV format
	if *outfile != "" {
		return runExternalParserMode()
	}

	// Legacy mode: use -f and -d flags
	var cdate time.Time
	var err error

	if *sdate != "" {
		cdate, err = time.Parse("2006-01-02", *sdate)
		if err != nil {
			return fmt.Errorf("failed to parse cutoff date %w", err)
		}
	} else {
		cdate = time.Now()
	}

	m, err := repo.GetNyhlMappings()
	if err != nil {
		return err
	}

	log.Println("date", cdate)
	importer := schimport.NewImporter(repo, cfg.ApiKey, cfg.ImportUrl)

	if *infile == "" {
		return importer.FetchAndImport("nyhl", m, cdate)
	}

	b, err := os.ReadFile(*infile)
	if err != nil {
		return err
	}

	var data schimport.Data
	if err = json.Unmarshal(b, &data); err != nil {
		return err
	}

	if len(data.Games) == 0 {
		log.Println("no games to import")
		return nil
	}

	var doc *html.Node
	switch path.Ext(*infile) {
	case ".json":
		return importer.ImportJson("nyhl", data, cdate, m)

	case ".xlx":
		b, err = os.ReadFile(*infile)
		if err != nil {
			return fmt.Errorf("failed to read file %s, %w", *infile, err)
		}

		// convert utf16 to utf8
		data, _, _ := transform.Bytes(unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder(), b)

		doc, err = htmlquery.Parse(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to read file %s, %w", *infile, err)
		}

		err = importer.Importxls("nyhl", doc, cdate, m)
	}
	return err

}
