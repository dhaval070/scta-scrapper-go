#!/bin/bash
ulimit -n 50000
dir="var/"`date +%Y%m%d`
d1="$dir/with-surface"

mkdir -p $d1

for site in `ls cmd/sites/`; do
    echo "$site"

    go run ./cmd/sites/$site/ --import-locations --outfile /tmp/$site.csv
    if [ $? -ne 0 ]; then
        echo "$site failed"
        exit
    fi

    f="$dir/$site.csv"
    cat /tmp/$site.csv | cut -d',' -f1-6 > $f
    if [ $? -ne 0 ]; then
        echo "failed $site csv process"
        exit
    fi

    go run ./cmd/site-schedule/main.go -site $site -infile $f > $d1/$site.csv &
done
