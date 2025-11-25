package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"calendar-scrapper/config"
	"calendar-scrapper/internal/schimport"
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
	cfg    config.Config
	repo   *repository.Repository
	infile *string
	sdate  *string
)

func init() {
	config.Init("config", ".")
	cfg = config.MustReadConfig()
	repo = repository.NewRepository(cfg)

	infile = cmd.Flags().StringP("file", "f", "", "CSV file path (required)")
	sdate = cmd.Flags().StringP("cutoffdate", "d", "", "date-from to import events (required) . e.g. -cutoffdate 2023-01-01")
}

func main() {
	cmd.Execute()
}

func runNyhl() error {
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
