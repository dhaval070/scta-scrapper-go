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

mmyyyy="${sd:4:2}${sd:0:4}"
echo "$site $mmyyyy"

go run ./cmd/sites/$site/ --import-locations -date $mmyyyy --outfile /tmp/$site.csv

if [ $? -ne 0 ]; then
    echo "$site failed"
    exit
fi

f="$dir/$site.csv"
csvtool -c 1-6 /tmp/$site.csv > $f

go run ./cmd/site-schedule/main.go -site $site -infile $f > $d1/$site.csv --import -cutoffdate $dt &
