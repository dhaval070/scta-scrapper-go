package main

import (
	"calendar-scrapper/pkg/parser"
	"encoding/csv"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"flag"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func parse(doc *html.Node, intdt int) {

}

func main() {
	var err error
	ymd := time.Now().Format("20060102")

	today := flag.String("today", ymd, "parse from date(yyyymmdd)")
	outfile := flag.String("outfile", "", "output filename")

	flag.Parse()

	log.Println(*today)

	intdt, err := strconv.Atoi(*today)
	if err != nil {
		log.Fatal(err)
	}

	var sites = []string{
		"https://niagaradistricthockeyleague.com/Calendar/",
		"https://srll.ca/Calendar/",
		"https://ysmhl.net/Calendar/",
		"https://tcmhl.ca/Calendar/",
		"https://lakeshorehockeyleague.net/Calendar/",
	}

	var results = make([][]string, 0)
	var wg sync.WaitGroup
	var m = sync.Mutex{}

	re := regexp.MustCompile(`^(\w+).*$`)

	for _, s := range sites {
		wg.Add(1)

		go func(s string, m *sync.Mutex) {
			doc, err := htmlquery.LoadURL(s)
			if err != nil {
				log.Println(err)
				return
			}

			result := parser.ParseSchedules(doc, intdt)
			log.Println("done ", s, len(result))

			for _, r := range result {
				r[5] = re.ReplaceAllString(r[5], "$1")
			}
			m.Lock()
			results = append(results, result...)
			m.Unlock()
			wg.Done()
		}(s, &m)
	}

	wg.Wait()

	sort.Sort(parser.ByDate(results))

	log.Println("total events", len(results))

	file, err := os.Create(*outfile)
	if err != nil {
		panic(err)
	}

	wr := csv.NewWriter(file)
	wr.WriteAll(results)

	if err = file.Close(); err != nil {
		panic(err)
	}
}
