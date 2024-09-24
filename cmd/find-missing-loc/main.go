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
// ID
// Full Name
// Full Short Name
// Street
// City
// Province
// cameraURL

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

	wr.Write([]string{"line#", "Error", "ID", "Full Name", "Full Short Name", "Street", "city", "Province", "camurl "})

	config.Init("config", ".")
	cfg := config.MustReadConfig()
	db, err := gorm.Open(mysql.Open(cfg.DbDSN))

	if err != nil {
		panic(err)
	}

	reader := csv.NewReader(fh)
	_, _ = reader.Read()

	reg := regexp.MustCompile("[^a-zA-Z0-9\\s]")

	for i := 2; ; i += 1 {
		si := fmt.Sprintf("%d", i)
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			panic(err)
		}

		var loc = &model.Location{}

		err = db.Raw(`select * from locations where city like ?`, "%"+row[4]+"%").Scan(loc).Error
		if err != nil {
			panic(err)
		}

		if loc.ID == 0 {
			fmt.Println(i, "city "+row[4]+" not found", row)
			r := append([]string{si, "city " + row[4] + " not found"}, row...)
			wr.Write(r)
			continue
		}
		// fmt.Println(row)
		street := reg.ReplaceAllString(row[3], "")

		err = db.Raw(`select * from locations where address1 like ?`, "%"+street+"%").Scan(loc).Error
		if err != nil {
			panic(err)
		}

		if loc.ID != 0 {
			if loc.TotalSurfaces == 0 {
				fmt.Println(i, "no surface", row)
				r := append([]string{si, "no surface"}, row...)
				wr.Write(r)
			}
			continue
		}

		var st = street
		parts := strings.Split(street, " ")
		if len(parts) > 3 {
			st := strings.Join(parts[:4], " ")

			err = db.Raw(`select * from locations where address1 like ?`, "%"+st+"%").Scan(loc).Error
			if err != nil {
				panic(err)
			}

			if loc.ID != 0 {
				if loc.TotalSurfaces == 0 {
					fmt.Println(i, "no surface", row)
					r := append([]string{si, "no surface"}, row...)
					wr.Write(r)
				}
				continue
			}

		}

		st = strings.Replace(st, "Road", "Rd", -1)
		err = db.Raw(`select * from locations where address1 like ?`, "%"+st+"%").Scan(loc).Error
		if err != nil {
			panic(err)
		}

		if loc.ID != 0 {
			if loc.TotalSurfaces == 0 {
				fmt.Println(i, "no surface", row)
				r := append([]string{si, "no surface"}, row...)
				wr.Write(r)
			}
			continue
		}

		st = strings.Replace(st, "Drive", "Dr", -1)
		err = db.Raw(`select * from locations where address1 like ?`, "%"+st+"%").Scan(loc).Error
		if err != nil {
			panic(err)
		}

		if loc.ID != 0 {
			if loc.TotalSurfaces == 0 {
				fmt.Println(i, "no surface", row)
				r := append([]string{si, "no surface"}, row...)
				wr.Write(r)
			}
			continue
		}

		r := append([]string{si, "address " + row[3] + " not found"}, row...)
		wr.Write(r)
		fmt.Println(i, "address not found", row)
	}
}
