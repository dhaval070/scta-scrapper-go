package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"calendar-scrapper/config"
	"calendar-scrapper/internal/schimport"
	"calendar-scrapper/pkg/repository"

	"github.com/antchfx/htmlquery"
	"github.com/spf13/cobra"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var cmd = &cobra.Command{
	Use:   "gthl-import",
	Short: "Import gthl schedule",
	RunE: func(c *cobra.Command, args []string) error {
		return runGthl()
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

	infile = cmd.Flags().StringP("file", "f", "", "XLS or json file path (required)")
	sdate = cmd.Flags().StringP("cutoffdate", "d", "", "date-from to import events (required) . e.g. -cutoffdate 2023-01-01")

	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("cutoffdate")
}

func main() {
	cmd.Execute()
}

func detectContentCharset(body io.Reader) string {
	r := bufio.NewReader(body)
	if data, err := r.Peek(1024); err == nil {
		if _, name, ok := charset.DetermineEncoding(data, ""); ok {
			return name
		}
	}
	return "utf-8"
}

func runGthl() error {
	var cdate time.Time
	var err error
	cdate, err = time.Parse("2006-01-02", *sdate)

	if err != nil {
		return fmt.Errorf("failed to parse cutoff date %w", err)
	}

	m, err := repo.GetGthlMappings()
	if err != nil {
		return err
	}

	switch path.Ext(*infile) {
	case ".json":
		return schimport.ImportJson(repo, "gthl", *infile, cdate, m)
	case ".xlx":
		b, err := os.ReadFile(*infile)
		if err != nil {
			return fmt.Errorf("failed to read file %s, %w", *infile, err)
		}

		// convert utf16 to utf8
		data, _, _ := transform.Bytes(unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder(), b)

		doc, err := htmlquery.Parse(bytes.NewReader(data))

		if err != nil {
			return fmt.Errorf("failed to read file %s, %w", *infile, err)
		}

		return schimport.Importxls(repo, "gthl", doc, cdate, m)
	}
	return errors.New("invalid file format")
}
