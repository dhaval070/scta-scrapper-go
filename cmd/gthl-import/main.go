package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"calendar-scrapper/pkg/repository"

	"github.com/antchfx/htmlquery"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
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

	infile = cmd.Flags().StringP("file", "f", "", "XLS file path (required)")
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

	b, err := os.ReadFile(*infile)
	if err != nil {
		return fmt.Errorf("failed to read file %s, %w", *infile, err)
	}

	// fmt.Println(detectContentCharset(bytes.NewReader(b)))
	// convert utf16 to utf8
	data, _, _ := transform.Bytes(unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder(), b)

	doc, err := htmlquery.Parse(bytes.NewReader(data))
	// log.Println(doc)

	if err != nil {
		return fmt.Errorf("failed to read file %s, %w", *infile, err)
	}

	err = importEvents(doc, cdate, m)
	return err
}

// Game_id	Rink	StartTime	StartDate	Division	Category	Visitor	VisitorTeamID	Home	HomeTeamID	GameType
func importEvents(root *html.Node, cutOffDate time.Time, mapping map[string]int) error {
	var err error
	var SourceType = "file"

	m := []*model.Event{}

	rows, err := htmlquery.QueryAll(root, "//table/tbody/tr")
	if err != nil {
		return err
	}

	ww := csv.NewWriter(os.Stdout)

	var r = make([]string, 11)

	for _, row := range rows[1:] {
		cols, err := htmlquery.QueryAll(row, "//td")
		if err != nil {
			return err
		}

		dt, err := parseDate(htmlquery.InnerText(cols[3]), htmlquery.InnerText(cols[2]))
		if err != nil {
			return err
		}

		if dt.Before(cutOffDate) {
			continue
		}

		sid, ok := mapping[htmlquery.InnerText(cols[1])]
		if !ok {
			log.Printf("failed to map surfaceid %s\n", htmlquery.InnerText(cols[1]))
			continue
		}

		m = append(m, &model.Event{
			Site:        "gthl",
			SourceType:  SourceType,
			Datetime:    dt,
			HomeTeam:    htmlquery.InnerText(cols[8]),
			OidHome:     htmlquery.InnerText(cols[9]),
			GuestTeam:   htmlquery.InnerText(cols[6]),
			OidGuest:    htmlquery.InnerText(cols[7]),
			Location:    htmlquery.InnerText(cols[1]),
			Division:    htmlquery.InnerText(cols[4]) + " " + htmlquery.InnerText(cols[5]),
			SurfaceID:   int32(sid),
			DateCreated: time.Now(),
		})

		for i := 0; i < 10; i += 1 {
			r[i] = htmlquery.InnerText(cols[i])
		}
		r[3] = dt.Format("2006-01-02")
		r[10] = fmt.Sprint(sid)

		ww.Write(r)
	}

	log.Println("total events ", len(m))
	err = repo.ImportEvents("gthl", m, cutOffDate)
	return err
}

func parseDate(date, t string) (tt time.Time, err error) {
	t1, err := time.Parse("15:04", t)
	if err != nil {
		return tt, fmt.Errorf("failed to parse time:%s %w", t, err)
	}

	// 01-Oct-2023
	dt, err := time.Parse("02-Jan-2006 15:04:05", date+" "+t1.Format("15:04:05"))

	if err != nil {
		return tt, fmt.Errorf("failed to parse date:%s time:%s :%w", date, t, err)
	}
	return dt, nil
}
