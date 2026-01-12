#!/bin/bash
go build -o ./bin/atlantichockeyfederation ./cmd/sites/atlantichockeyfederation
go build -o ./bin/alliancehockey ./cmd/sites/alliancehockey
go build -o ./bin/lugsports ./cmd/sites/lugsports
go build -o ./bin/gamesheet ./cmd/sites/gamesheet
go build -o ./bin/csv-trim-last ./cmd/csv-trim-last
go build -o ./bin/site-schedule ./cmd/site-schedule
go build ./cmd/scraper/
