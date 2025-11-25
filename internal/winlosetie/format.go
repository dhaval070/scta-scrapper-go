package winlosetie

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

var ErrUnsupportedSite = errors.New("Unsupported site")
var barnUrl = "https://livebarn.com/en/video/%s/%s/%s"

type DataRec struct {
	Id        int
	Date      string
	Time      string
	SurfaceID string
}

func formatInput(r io.Reader, site string) ([]DataRec, error) {
	switch site {
	case "":
		return unformatted(r)
	case "nyhl":
		return formatNyhl(r)
	default:
		return nil, ErrUnsupportedSite
	}
}

// id, date, time, surface id
// 793537,2023-11-11,14:10,3876
func unformatted(r io.Reader) ([]DataRec, error) {
	csvr := csv.NewReader(r)
	data := []DataRec{}
	for {
		rec, err := csvr.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}
		id, err := strconv.Atoi(rec[0])
		if err != nil {
			return nil, fmt.Errorf("id %s is not a number. %w", rec[0], err)
		}

		data = append(data, DataRec{
			Id:        id,
			Date:      rec[1],
			Time:      rec[2],
			SurfaceID: rec[3],
		})
	}
	return data, nil
}

// NYHL format required: GameID	League	Season	Division	Tier	group HomeTeam	Tier group	VisitorTeam	Location	l
func formatNyhl(r io.Reader) (data []DataRec, err error) {
	data = make([]DataRec, 0)
	csvr := csv.NewReader(r)
	for {
		rec, err := csvr.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("csv read error %w", err)
		}
		id, err := strconv.Atoi(rec[0])
		if err != nil {
			return nil, fmt.Errorf("id %s is not a number. %w", rec[0], err)
		}

		dt, err := time.Parse("1/2/2006", rec[11])
		if err != nil {
			return nil, err
		}

		tt, err := time.Parse("3:04 PM", rec[12])
		if err != nil {
			return nil, fmt.Errorf("failed to parse time %w", err)
		}

		data = append(data, DataRec{
			Id:        id,
			Date:      dt.Format("2006-01-02"),
			Time:      tt.Format("15:04"),
			SurfaceID: rec[13],
		})
	}

	return data, err
}
