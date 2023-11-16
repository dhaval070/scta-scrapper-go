# calendar-scrapper

## Import master locations
Download JSON file from https://staticdata.livebarn.com/api/v2.0.0/staticdata/venues
```
wget https://staticdata.livebarn.com/api/v2.0.0/staticdata/venues -O master-locations.json
```

and run following command:

```
go run ./cmd/venue-import/ --path=master-locations.json
```


## Run all the sites scrappers
```
sh run-all.sh
```
It will scrape all the sites schedules with addresses, import new locations, run matches and generate csv files under ```./var/<current date>/with-surface/``` directory
