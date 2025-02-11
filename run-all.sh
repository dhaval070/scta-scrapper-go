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

for site in `ls cmd/sites/`; do
    echo "$site $mmyyyy"

    go run ./cmd/sites/$site/ --import-locations -date $mmyyyy --outfile /tmp/$site.csv
    if [ $? -ne 0 ]; then
        echo "$site failed"
        exit
    fi

    f="$dir/$site.csv"
    csvtool -c 1-6 /tmp/$site.csv > $f

    go run ./cmd/site-schedule/main.go -site $site -infile $f > $d1/$site.csv --import -cutoffdate $dt &
done
