package writer

import (
	"encoding/csv"
	"io"
)

func WriteCsv(fh io.WriteCloser, data [][]string) {

	csv := csv.NewWriter(fh)

	for _, row := range data {
		//fmt.Fprintln(fh, strings.Join(row, ","))
		if err := csv.Write(row); err != nil {
			panic(err)
		}
	}
	csv.Flush()

	if err := csv.Error(); err != nil {
		panic(err)
	}
	fh.Close()
}
