package writer

import (
	"encoding/csv"
	"io"
)

func WriteCsv(fh io.WriteCloser, data [][]string) {

	csv := csv.NewWriter(fh)
	csv.WriteAll(data)
	csv.Flush()

	if err := csv.Error(); err != nil {
		panic(err)
	}
	fh.Close()
}
