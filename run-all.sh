#!/bin/bash

for site in "alliancehockey" \
"beechey" \
"intertownll" \
"threecountyhockey" \
"eomhl" \
"ndll" \
"omha-aaa" \
"lmll" \
"victoriadurham" \
"essexll" \
"srll" \
"southerncounties" \
"leohockey" \
"haldimandll" \
"gbmhl" \
"gbtll" \
"fourcountieshockey" \
"woaa.on" \
"grandriverll" \
"shamrockhockey" \
"bluewaterhockey" \
"niagrahockey" \
"lakeshorehockey" \
"ysmhl" \
"tcmhl"; do
    echo "$site"
    #mysql --port=5506 -uapp -h 172.17.0.1 -papp schedules -e "insert into sites(site, url) values('$site','')"

    go run ./cmd/$site/ --import-locations --outfile var/15102023/$site.csv
    if [ $? -ne 0 ]; then
        exit
    fi
done
