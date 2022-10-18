package writer

import (
	"encoding/csv"
	"io"
)

func WriteCsv(fh io.WriteCloser, data [][]string) {

	csv := csv.NewWriter(fh)

	csv.Write([]string{"date", "division", "home team", "guest team", "location"})

	for _, row := range data {
		csv.Write(row)
	}
	fh.Close()
}
