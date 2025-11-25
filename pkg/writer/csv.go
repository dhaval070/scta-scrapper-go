package writer

import (
	"calendar-scrapper/dao/model"
	"encoding/csv"
	"fmt"
	"io"
)

func WriteCsv(fh io.WriteCloser, data [][]string) error {
	var err error
	csv := csv.NewWriter(fh)

	if err = csv.WriteAll(data); err != nil {
		return err
	}

	csv.Flush()

	if err := csv.Error(); err != nil {
		return err
	}
	return nil
}

func WriteEvents(w io.Writer, data []*model.Event) error {
	var err error
	ww := csv.NewWriter(w)
	for _, rec := range data {
		row := []string{
			rec.Datetime.Format("1/2/2006 15:04"),
			rec.HomeTeam,
			rec.GuestTeam,
			rec.Location,
			rec.Division,
			fmt.Sprint(rec.SurfaceID),
		}
		if err = ww.Write(row); err != nil {
			return err
		}
	}
	return nil
}
