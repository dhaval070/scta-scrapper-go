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
	"time"

	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/internal/schimport"
	"calendar-scrapper/pkg/repository"

	"github.com/antchfx/htmlquery"
	"github.com/spf13/cobra"
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
	}

	m, err := repo.GetNyhlMappings()
	if err != nil {
		return err
	}

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

	switch path.Ext(*infile) {
	case ".json":
		return importer.ImportJson("nyhl", data, cdate, m)

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

		err = importer.Importxls("nyhl", doc, cdate, m)
	}
	return err

}

// format required: GameID	League	Season	Division	Tier	group HomeTeam	Tier group	VisitorTeam	Location	Date	Time
func importEvents(ff *csv.Reader, cutOffDate time.Time, mapping map[string]int) error {
	var err error
	var SourceType = "file"

	m := []*model.Event{}

	ww := csv.NewWriter(os.Stdout)

	var r = make([]string, 14)

	for i := 1; ; i += 1 {
		cols, err := ff.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error at row %d, %w", i, err)
		}

		sid, ok := mapping[cols[10]]
		if !ok {
			log.Printf("failed to map surfaceid %s\n", cols[10])
			continue
		}

		dt, err := parseDate(cols[11], cols[12])
		if err != nil {
			return err
		}

		if dt.Before(cutOffDate) {
			continue
		}

		m = append(m, &model.Event{
			Site:        "nyhl",
			SourceType:  SourceType,
			Datetime:    dt,
			HomeTeam:    cols[6],
			GuestTeam:   cols[9],
			Location:    cols[10],
			Division:    cols[3],
			SurfaceID:   int32(sid),
			DateCreated: time.Now(),
		})

		for i := 0; i < 13; i += 1 {
			r[i] = cols[i]
		}
		r[13] = fmt.Sprint(sid)

		ww.Write(r)
	}

	ww.Flush()
	log.Println("total events ", len(m))
	err = repo.ImportEvents("nyhl", m, cutOffDate)
	return err
}

func parseDate(date, t string) (tt time.Time, err error) {
	t1, err := time.Parse("3:04 PM", t)
	if err != nil {
		return tt, fmt.Errorf("failed to parse time:%s %w", t, err)
	}

	dt, err := time.Parse("1/2/2006 15:04:05", date+" "+t1.Format("15:04:05"))
	if err != nil {
		return tt, fmt.Errorf("failed to parse date:%s time:%s %w", date, t, err)
	}
	return dt, nil
}
