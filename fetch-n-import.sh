dt=`date +%Y-%m-%d`
echo "cutoff date: $dt"

for site in "gthl" "nyhl" "mhl" ; do
    echo "processing $site"
    curl -H"x-api-key:2D0B32C3-900D-487B-89C5-E8D9012A559B" "https://www.agilex.ca/api/GameQuery?date=02-Oct-2024&id=AS0122341&league=$site" > $site.json

    go run ./cmd/$site-import -f $site.json -d $dt >$site.log  2>&1
done
