.PHONY: all build buildapi test fmt vet
.PHONY: atlantichockeyfederation alliancehockey lugsports gamesheet
.PHONY: csv-trim-last site-schedule agilex claim-api-test claim-cron scraper migrate

all: build buildapi

build: atlantichockeyfederation alliancehockey lugsports gamesheet \
       csv-trim-last site-schedule agilex claim-api-test claim-cron scraper migrate

buildapi:
	cd api && go build -ldflags='-s -w' -o ./surface-api .

atlantichockeyfederation:
	go build -o ./bin/atlantichockeyfederation ./cmd/sites/atlantichockeyfederation

alliancehockey:
	go build -o ./bin/alliancehockey ./cmd/sites/alliancehockey

lugsports:
	go build -o ./bin/lugsports ./cmd/sites/lugsports

gamesheet:
	go build -o ./bin/gamesheet ./cmd/sites/gamesheet

csv-trim-last:
	go build -o ./bin/csv-trim-last ./cmd/csv-trim-last

site-schedule:
	go build -o ./bin/site-schedule ./cmd/site-schedule

agilex:
	go build -o ./bin/agilex ./cmd/agilex

claim-api-test:
	go build -o ./bin/claim-api-test ./cmd/claim-api-test

claim-cron:
	go build -o ./bin/claim-cron ./cmd/claim-cron

scraper:
	go build ./cmd/scraper/

migrate:
	go build -ldflags='-s -w' -o bin/migrate ./cmd/migrate

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

-include deploy.env

deploy:
	@ssh $(DEPLOY_HOST) "set -e && \
	  export PATH=\$$PATH:/usr/local/go/bin && \
	  echo '==> Pulling latest code...' && \
	  cd $(DEPLOY_REMOTE_DIR) && \
	  git pull && \
	  echo '==> Building migrate binary...' && \
	  go build -ldflags='-s -w' -o ./bin/migrate ./cmd/migrate && \
	  echo '==> Running migrations...' && \
	  ./bin/migrate up && \
	  echo '==> Building surface-api...' && \
	  cd api && \
	  go build -ldflags='-s -w' -o ./surface-api . && \
	  echo '==> Stopping service...' && \
	  sudo systemctl stop surface-api && \
	  echo '==> Copying binary...' && \
	  cp ./surface-api $(DEPLOY_TARGET_DIR)/ && \
	  echo '==> Starting service...' && \
	  sudo systemctl start surface-api && \
	  echo '==> Done.'"
