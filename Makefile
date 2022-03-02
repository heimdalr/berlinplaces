SHELL = /bin/bash

# Just to be sure, add the path of the binary-based go installation.
PATH := /usr/local/go/bin:$(PATH)

# Using the (above extended) path, query the GOPATH (i.e. the user's go path).
GOPATH := $(shell env PATH=$(PATH) go env GOPATH)

# Add $GOPATH/bin to path
PATH := $(GOPATH)/bin:$(PATH)

# extract git hash
BUILD_GIT_HASH := $(shell git rev-parse HEAD 2>/dev/null || echo "0")
GIT_TAG := $(shell git describe --tags 2>/dev/null || echo "v0.0.0")
BUILD_VERSION := $(shell echo ${GIT_TAG} | grep -P -o '(?<=v)[0-9]+.[0-9]+.[0-9]')

# default golang flags
LD_FLAGS := '-X main.buildVersion=$(BUILD_VERSION) -X main.buildGitHash=$(BUILD_GIT_HASH)'

DB_FILE := vbb.db

# directory of Makefile
MAKEFILE_DIR = $(shell pwd)

GO_FILES := $(wildcard ./*.go)
BIN_FILES := syclist

all: $(BIN_FILES)

fmt:
	gofmt -s -w .

test:
	go test -p 4 -v ./...

lint:
	golangci-lint run

coverage:
	go test -coverprofile=c.out && go tool cover -html=c.out

syclist: $(GO_FILES)
	go build \
	-ldflags $(LD_FLAGS) \
	-o $@ \
	.

run-syclist: syclist
	SYCLIST_DB=$(DB_FILE) ./syclist

start_nominatim: stop_nominatim
	docker run --rm --shm-size=8g \
	-e PBF_URL=https://download.geofabrik.de/europe/germany/berlin-latest.osm.pbf \
	-e REPLICATION_URL=https://download.geofabrik.de/europe/germany/berlin-updates/ \
	-e IMPORT_WIKIPEDIA=false \
	-e FREEZE=true \
	-v /home/sebastian/blub/repos/private/geocoder/placechecker/nominatim-data:/var/lib/postgresql/12/main \
	-p 8081:8080 \
	-p 5432:5432 \
	--name nominatim \
	mediagis/nominatim:4.0

stop_nominatim:
	docker stop nominatim 2>/dev/null || true

docker_image: Dockerfile
	docker build -t syclist .

run_image: stop_image
	docker run -p 8080:8080 --name syclist syclist

stop_image:
	docker stop syclist 2>/dev/null || true
	docker rm syclist 2>/dev/null || true

create_db:
	go install github.com/heimdalr/gtfs/cmd/gtfs@latest
	rm -Rf /tmp/vbb
	mkdir /tmp/vbb
	curl -o /tmp/vbb/GTFS.zip https://www.vbb.de/fileadmin/user_upload/VBB/Dokumente/API-Datensaetze/gtfs-mastscharf/GTFS.zip
	unzip -d /tmp/vbb /tmp/vbb/GTFS.zip
	gtfs import /tmp/vbb $(DB_FILE)
	gtfs trim $(DB_FILE) "S-Bahn"
	cat cwt.sql | sqlite3 $(DB_FILE)

clean:
	rm -f $(BIN_FILES)

.PHONY: all fmt test lint coverage run-syclist clean
