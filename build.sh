#!/bin/bash
go build -o ./bin/alliancehockey ./cmd/sites/alliancehockey
go build -o ./bin/lugsports ./cmd/sites/lugsports
go build -o ./bin/gamesheet ./cmd/sites/gamesheet
go build ./cmd/scraper/
