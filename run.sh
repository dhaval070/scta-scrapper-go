#!/bin/bash
# usage: sh run.sh alliancehockey 2024-12-01
# date is optional, if not provided then current date is used
dt="$2"

if [ -z $dt ]; then
    dt=`date +%Y-%m-%d`
fi

sd=`echo $dt|sed 's/-//g'`
dir="var/$sd"
d1="$dir/with-surface"

mkdir -p $d1
site="$1"

yyyy=`echo $sd|cut -c 1-4`
mm=`echo $sd|cut -c 5-6`

mmyyyy="$mm$yyyy"
echo "$site $mmyyyy"

# Use new universal scraper instead of individual site binary
./scraper --site=$site --import-locations --date $mmyyyy --outfile /tmp/$site.csv

if [ $? -ne 0 ]; then
    echo "$site failed"
    exit
fi

if [ -e /tmp/${site}_/tmp/$site.csv -a -s /tmp/${site}_/tmp/$site.csv ]; then
    f="$dir/$site.csv"
    csvtool --encoding utf8 -c 1-6 /tmp/${site}_/tmp/$site.csv > $f
    
    go run ./cmd/site-schedule/main.go -site $site -infile $f > $d1/$site.csv --import -cutoffdate $dt &
fi
