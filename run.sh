#!/bin/bash
# usage: sh run.sh <sites> [date] [workers]
# sites: mandatory - can be --all or comma-separated list (e.g., site1,site2,site3)
# date: optional - defaults to current date (format: YYYY-MM-DD)
# workers: optional - number of concurrent workers (default: uses scraper default)
# Examples:
#   sh run.sh --all              # all sites, current date
#   sh run.sh site1,site2        # specific sites, current date
#   sh run.sh --all 2024-12-01   # all sites, specific date
#   sh run.sh site1 2024-12-01   # specific site, specific date
#   sh run.sh --all 2024-12-01 4 # all sites, specific date, 4 workers

# Sites is mandatory
sites="$1"
if [ -z "$sites" ]; then
    echo "Error: sites parameter is required"
    echo "Usage: sh run.sh <sites> [date] [workers]"
    echo "  sites: --all or comma-separated list (e.g., site1,site2,site3)"
    echo "  date: optional, defaults to current date (YYYY-MM-DD)"
    echo "  workers: optional, number of concurrent workers"
    exit 1
fi

if [ "$sites" != "--all" ]; then
    sites="--sites $sites"
fi
# Date is optional, defaults to current date
dt="$2"
if [ -z "$dt" ]; then
    dt=`date +%Y-%m-%d`
fi

# Workers is optional
workers="$3"
workers_flag=""
if [ -n "$workers" ]; then
    workers_flag="--workers $workers"
fi

sd=`echo $dt|sed 's/-//g'`

dt=`date +%Y-%m-%d`
dir="var/$sd"
d1="$dir/with-surface"

yyyy=`echo $sd|cut -c 1-4`
mm=`echo $sd|cut -c 5-6`

mmyyyy="$mm$yyyy"

mkdir -p $d1

# Create unique temporary directory
tmpdir=$(mktemp -d /tmp/calendar-scraper.XXXXXX)

# Use new universal scraper instead of individual site binaries
echo "Running universal scraper for sites: $sites with date: $mmyyyy"

./scraper $sites --import-locations --date $mmyyyy --outfile $tmpdir/schedules.csv $workers_flag

if [ $? -ne 0 ]; then
    echo "Scraper failed"
    exit 1
fi

# Process each site's output file
for csv_file in $tmpdir/*_schedules.csv; do
    if [ -e "$csv_file" ]; then
        site=$(basename "$csv_file" _schedules.csv)
        echo "Processing $site"

        f="$dir/$site.csv"
        csvtool --encoding utf8 -c 1-6 "$csv_file" > $f

        go run ./cmd/site-schedule/main.go -site $site -infile $f > $d1/$site.csv --import -cutoffdate $dt &
    fi
done

# Clean up temporary directory
rm -rf $tmpdir
