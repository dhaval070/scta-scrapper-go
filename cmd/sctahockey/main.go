package main

import (
	"calendar-scrapper/pkg/parser"
	"calendar-scrapper/pkg/writer"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"flag"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func main() {
	ymd := time.Now().Format("20060102")

	infile := flag.String("infile", "", "local html filename")
	today := flag.String("today", ymd, "parse from date(yyyymmdd)")
	outfile := flag.String("outfile", "", "output filename")

	flag.Parse()

	log.Println(*today)
	log.Println(*infile)

	var doc *html.Node
	var err error

	if *infile != "" {
		doc, err = htmlquery.LoadDoc(*infile)
	} else {
		doc, err = htmlquery.LoadURL("https://sctahockey.com/Calendar/")
	}

	if err != nil {
		log.Fatal(err)
	}

	intdt, err := strconv.Atoi(*today)
	if err != nil {
		log.Fatal(err)
	}

	result := parser.ParseSchedules(doc, intdt)
	sort.Sort(parser.ByDate(result))

	if *outfile != "" {
		fh, err := os.Create(*outfile)
		if err != nil {
			log.Fatal(err)
		}
		writer.WriteCsv(fh, result)
	} else {
		log.Println(result)
	}

}
