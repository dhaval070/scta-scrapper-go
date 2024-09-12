package main

import (
	"calendar-scrapper/pkg/parser"
	"os"

	"github.com/antchfx/htmlquery"
)

func main() {
	fh, err := os.Open("/home/dhaval/rust/httpdemo/data/eomhl.html")
	if err != nil {
		panic(err)
	}

	for i := 0; i < 100; i += 1 {
		fh.Seek(0, 0)
		doc, err := htmlquery.Parse(fh)
		if err != nil {
			panic(err)
		}

		_ = parser.ParseSchedules("eomhl", doc)
	}
}
