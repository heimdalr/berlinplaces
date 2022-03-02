# use the latest (debian based) golang base image for building
FROM golang:bullseye AS buildstage

# update and get needed packages
RUN apt-get update
RUN apt-get install -y sqlite3 unzip

# download the gtfs tool (see below)
RUN go install github.com/heimdalr/gtfs/cmd/gtfs@latest

# download and unpack VBB data
WORKDIR /vbb
RUN wget https://www.vbb.de/fileadmin/user_upload/VBB/Dokumente/API-Datensaetze/gtfs-mastscharf/GTFS.zip
RUN unzip GTFS.zip

# mangle the VBB data into a SQLite DB
WORKDIR /
RUN gtfs import /vbb vbb.db
RUN gtfs trim vbb.db "S-Bahn"

# mangle the VBB DB to suite syclist needs
ADD cwt.sql .
RUN cat cwt.sql | sqlite3 /vbb.db

# download syclist dependencies
WORKDIR /syclist
COPY go.mod ./
COPY go.sum ./
RUN go mod download
ADD *.go ./

# add the .git directory as we do versioning based on hash and tag
ADD .git ./.git

# build the syclist
RUN export BUILD_GIT_HASH=$(git rev-parse HEAD 2>/dev/null || echo '0') && \
    export GIT_TAG=$(git describe --tags 2>/dev/null || echo 'v0.0.0') && \
    export BUILD_VERSION=$(echo $GIT_TAG | grep -P -o '(?<=v)[0-9]+.[0-9]+.[0-9]') && \
    go build \
    -ldflags "-X main.buildVersion=$BUILD_VERSION -X main.buildGitHash=$BUILD_GIT_HASH" \
    -o syclist \
    .

# use debian for running (sqlite needs CGO which contradicts alpine and busybox)
FROM debian

# copy executable and DB from the buildstage
COPY --from=buildstage /syclist/syclist /
COPY --from=buildstage /vbb.db /

# add statics
ADD swagger /swagger
ADD web /web


ENV SYCLIST_DB=/vbb.db
ENV SYCLIST_MODE=release
#ENV SYCLIST_GEOAPIFY_KEY=XYZ

ENTRYPOINT [ "/syclist" ]
