package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	infile := flag.String("infile", "", "input CSV file (required)")
	outfile := flag.String("outfile", "", "output CSV file (defaults to stdout)")
	flag.Parse()

	if *infile == "" {
		fmt.Fprintln(os.Stderr, "Error: -infile is required")
		os.Exit(1)
	}

	in, err := os.Open(*infile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer in.Close()

	reader := csv.NewReader(in)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	var outW *csv.Writer
	if *outfile != "" {
		out, err := os.Create(*outfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer out.Close()
		outW = csv.NewWriter(out)
	} else {
		outW = csv.NewWriter(os.Stdout)
	}
	defer outW.Flush()

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading CSV: %v\n", err)
			os.Exit(1)
		}

		if len(rec) > 0 {
			if err := outW.Write(rec[:len(rec)-1]); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Empty record; write as-is
			if err := outW.Write(rec); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
				os.Exit(1)
			}
		}
	}
}
