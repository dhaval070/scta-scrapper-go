#!/bin/bash
# usage: sh run-all.sh 2024-12-01
# date is optional, if not provided then current date is used
dt="$1"

if [ -z $dt ]; then
    dt=`date +%Y-%m-%d`
fi

sd=`echo $dt|sed 's/-//g'`

dt=`date +%Y-%m-%d`
dir="var/$sd"
d1="$dir/with-surface"

yyyy=`echo $sd|cut -c 1-4`
mm=`echo $sd|cut -c 5-6`

mmyyyy="$mm$yyyy"

mkdir -p $d1

# Use new universal scraper instead of individual site binaries
echo "Running universal scraper for all sites: $mmyyyy"

./scraper --all --import-locations --date $mmyyyy --outfile /tmp/schedules.csv

if [ $? -ne 0 ]; then
    echo "Scraper failed"
    exit 1
fi

# Process each site's output file
for csv_file in /tmp/*_schedules.csv; do
    if [ -e "$csv_file" ]; then
        site=$(basename "$csv_file" _schedules.csv)
        echo "Processing $site"

        f="$dir/$site.csv"
        csvtool --encoding utf8 -c 1-6 "$csv_file" > $f

        go run ./cmd/site-schedule/main.go -site $site -infile $f > $d1/$site.csv --import -cutoffdate $dt &
    fi
done
