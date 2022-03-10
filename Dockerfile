# use the latest (debian based) golang base image for building
FROM golang:bullseye AS buildstage

# update and get needed packages
#RUN apt-get update
#RUN apt-get install -y sqlite3 unzip

# download app dependencies
WORKDIR /builddir
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY main.go ./
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# add the .git directory as we do versioning based on hash and tag
ADD .git ./.git

# build the syclist
RUN export BUILD_GIT_HASH=$(git rev-parse HEAD 2>/dev/null || echo '0') && \
    export GIT_TAG=$(git describe --tags 2>/dev/null || echo 'v0.0.0') && \
    export BUILD_VERSION=$(echo $GIT_TAG | grep -P -o '(?<=v)[0-9]+.[0-9]+.[0-9]') && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X main.buildVersion=$BUILD_VERSION -X main.buildGitHash=$BUILD_GIT_HASH" \
    -o places \
    .

# create api user
# See https://stackoverflow.com/a/55757473/12429735RUN
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "10001" \
    "api"

# use scratch for the final image
#FROM scratch
FROM alpine

# copy ca certificates
COPY --from=buildstage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# import the user and group files from the builder
COPY --from=buildstage /etc/passwd /etc/passwd
COPY --from=buildstage /etc/group /etc/group


WORKDIR /places

# copy server
COPY --from=buildstage /builddir/places ./

# copy statics
COPY swagger/ ./swagger/
COPY web/ ./web/
COPY _data/districts.csv ./_data/districts.csv
COPY _data/streets.csv ./_data/streets.csv
COPY _data/locations.csv ./_data/locations.csv
COPY _data/housenumbers.csv ./_data/housenumbers.csv

# use api user
USER api:api

#ENV BP_CSV="berlin.csv"
#ENV BP_PORT=8080
ENV BP_MODE="release"

CMD [ "/places/places" ]
