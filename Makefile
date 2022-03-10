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

# directory of Makefile
MAKEFILE_DIR = $(shell pwd)

GO_FILES := $(wildcard ./pkg/places/*.go ./internal/*.go ./*.go)

all: berlinplaces

data: _data/extractCSV.sql _data/extractCSV.sh
	cd _data && ./extractCSV.sh

fmt:
	go fmt .
	go fmt github.com/heimdalr/berlinplaces/pkg/...

test:
	go test -p 4 -v ./...

lint:
	golangci-lint run

coverage:
	go test -coverprofile=c.out github.com/heimdalr/berlinplaces/pkg/... && go tool cover -html=c.out

berlinplaces: $(GO_FILES)
	go build \
	-ldflags $(LD_FLAGS) \
	-o $@ \
	.

run_berlinplaces: berlinplaces
	./berlinplaces

build_image: Dockerfile
	docker build -t berlinplaces .

run_image: stop_image
	docker run -p 8080:8080 --name berlinplaces berlinplaces

stop_image:
	docker stop berlinplaces 2>/dev/null || true
	docker rm berlinplaces 2>/dev/null || true

start_nominatim:
	docker run --rm -d --shm-size=8g \
        -e PBF_URL=https://download.geofabrik.de/europe/germany/berlin-latest.osm.pbf \
        -e REPLICATION_URL=https://download.geofabrik.de/europe/germany/berlin-updates/ \
        -e IMPORT_WIKIPEDIA=false \
        -e FREEZE=true \
        -v $(MAKEFILE_DIR)/_data/nominatim/.nominatim-data:/var/lib/postgresql/12/main \
        -p 8081:8080 \
        -p 5432:5432 \
        --name nominatim \
        mediagis/nominatim:4.0

stop_nominatim:
	docker stop nominatim 2>/dev/null || true

clean:
	rm -f berlinplaces c.out

.PHONY: all data fmt test lint coverage run_berlinplaces build_image run_image stop_image start_nominatim stop_nominatim clean
