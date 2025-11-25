package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"calendar-scrapper/config"
	"calendar-scrapper/pkg/repository"
)

var repo *repository.Repository

type Row struct {
	Datetime    string
	Site        string
	HomeTeam    string
	GuestTeam   string
	Location    string
	SurfaceId   int32
	SurfaceName string
}

type Schedules []Row

func (s Schedules) Len() int           { return len(s) }
func (s Schedules) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Schedules) Less(i, j int) bool { return s[i].Datetime < s[j].Datetime }

func processCsv(r io.Reader) {
	rr := csv.NewReader(r)

	var result = map[int32][]Row{}

	for {
		r, err := rr.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if len(r) < 4 {
			log.Fatalf("invalid columns %+v\n", r)
		}
		row := Row{r[0], r[1], r[2], r[3], r[4], 0, ""}

		surface := repo.GetMatchingSurface(row.Site, row.Location)
		if surface == nil {
			continue
		}

		row.SurfaceId = surface.ID
		row.SurfaceName = surface.Name

		result[surface.ID] = append(result[surface.ID], row)
	}

	ww := csv.NewWriter(os.Stdout)
	err := ww.Write([]string{
		"surface ID", "surface name", "date", "site", "home team", "GuestTeam", "Location",
	})
	if err != nil {
		panic(err)
	}

	for _, recs := range result {
		sort.Sort(Schedules(recs))

		for _, r := range recs {
			id := fmt.Sprint(r.SurfaceId)

			err := ww.Write([]string{
				id, r.SurfaceName, r.Datetime, r.Site, r.HomeTeam, r.GuestTeam, r.Location,
			})
			if err != nil {
				panic(err)
			}
		}
	}
	ww.Flush()
}

func main() {
	config.Init("config", ".")

	var cfg = config.MustReadConfig()
	repo = repository.NewRepository(cfg)

	infile := flag.String("infile", "", "schedule csv file")
	flag.Parse()

	if infile == nil {
		log.Fatal("infle is required")
	}

	files := strings.Split(*infile, ",")

	content := ""
	for _, fn := range files {
		c, err := os.ReadFile(fn)
		if err != nil {
			panic(err)
		}
		content = content + string(c)
	}

	r := strings.NewReader(content)
	// fh, err := os.Open(*infile)
	// if err != nil {
	// 	panic(err)
	// }

	processCsv(r)
}
