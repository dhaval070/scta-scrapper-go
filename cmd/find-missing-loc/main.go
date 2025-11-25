package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/dao/model"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// input csv fields:
//0 ID
//1 Full Name
//2 Full Short Name
//3 Street
//4 City
//5 Province
//6 cameraURL

func main() {
	fmt.Println("args", os.Args)

	fh, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	out, err := os.Create("missig-loc.csv")
	if err != nil {
		panic(err)
	}
	wr := csv.NewWriter(out)

	config.Init("config", ".")
	cfg := config.MustReadConfig()
	db, err := gorm.Open(mysql.Open(cfg.DbDSN))

	if err != nil {
		panic(err)
	}

	reader := csv.NewReader(fh)
	header, err := reader.Read()
	if err != nil {
		panic(err)
	}

	err = wr.Write(append(header, "surface ID", "Location ID"))
	if err != nil {
		panic(err)
	}

	reg := regexp.MustCompile(`[^a-zA-Z0-9\s]`)

	for i := 2; ; i += 1 {
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			panic(err)
		}

		var loc = &model.Location{}

		var city = row[4]
		err = db.Raw(`select * from locations where city like ?`, "%"+city+"%").Scan(loc).Error
		if err != nil {
			panic(err)
		}

		if loc.ID == 0 {
			if err := wr.Write(row); err != nil {
				panic(err)
			}
			continue
		}
		street := reg.ReplaceAllString(row[3], "")
		if street == "" {
			if err := wr.Write(row); err != nil {
				panic(err)
			}
			continue
		}

		loc = findByStreet(db, "%"+street+"%", city)

		if loc.ID != 0 {
			writeRow(db, row, loc.ID, wr)
			continue
		}

		var st = street
		parts := strings.Split(street, " ")
		if len(parts) > 3 {
			st := strings.Join(parts[:4], " ")

			loc = findByStreet(db, "%"+st+"%", city)

			if loc.ID != 0 {
				writeRow(db, row, loc.ID, wr)
				continue
			}
		}

		st = strings.Replace(st, "Road", "Rd", -1)
		loc = findByStreet(db, "%"+st+"%", city)

		if loc.ID != 0 {
			writeRow(db, row, loc.ID, wr)
			continue
		}

		st = strings.Replace(st, "Drive", "Dr", -1)
		loc = findByStreet(db, "%"+st+"%", city)

		if loc.ID != 0 {
			writeRow(db, row, loc.ID, wr)
			continue
		}

		err = wr.Write(row)
		if err != nil {
			panic(err)
		}
	}
}

func getSurfaceIDs(db *gorm.DB, locId int32) string {
	// db.Find()
	var ids = []string{}

	err := db.Raw("select id from surfaces where location_id=?", locId).Scan(&ids).Error
	if err != nil {
		panic(err)
	}

	return strings.Join(ids, ",")
}

func writeRow(db *gorm.DB, row []string, locID int32, wr *csv.Writer) {
	id := getSurfaceIDs(db, locID)
	row = append(row, id, fmt.Sprintf("%d", locID))
	err := wr.Write(row)
	if err != nil {
		panic(err)
	}
}

func findByStreet(db *gorm.DB, street string, city string) *model.Location {
	var loc = &model.Location{}

	city = "%" + city + "%"
	err := db.Raw(`select * from locations where address1 like ? and city like ?`, street, city).Scan(loc).Error
	if err != nil {
		panic(err)
	}

	return loc
}
