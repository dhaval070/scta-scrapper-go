#!/bin/bash

dir="var/16112023"
d1="$dir/with-surface"

mkdir -p $d1

for site in `ls cmd/sites/`; do
    echo "$site"
    #mysql --port=5506 -uapp -h 172.17.0.1 -papp schedules -e "insert into sites(site, url) values('$site','')"

    go run ./cmd/sites/$site/ --import-locations --outfile /tmp/$site.csv
    if [ $? -ne 0 ]; then
        exit
    fi

    f="$dir/$site.csv"
    cat /tmp/$site.csv | cut -d',' -f1,2,3,4,5,6 > $f
    if [ $? -ne 0 ]; then
        echo "failed $site"
        exit
    fi

    go run ./cmd/site-schedule/main.go -site $site -infile $f > $d1/$site.csv &
done
