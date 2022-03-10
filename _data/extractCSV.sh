#!/usr/bin/env bash

CONTAINER_NAME=nominatim
NOMINATIM_DIR=${PWD}/nominatim
NOMINATIM_DATA=${NOMINATIM_DIR}/.nominatim-data
CSV_FILE_DISTRICTS=${PWD}/districts.csv
CSV_FILE_STREETS=${PWD}/streets.csv
CSV_FILE_LOCATIONS=${PWD}/locations.csv
CSV_FILE_HOUSENUMBERS=${PWD}/housenumbers.csv
SQL_FILE=${PWD}/extractCSV.sql
CONTAINER_CHECK_URL=http://localhost:8081/search.php?q=Oranienburger

mkdir -p "${NOMINATIM_DIR}"

# if Nominatim container is not running
if ! docker ps --format '{{.Names}}' | grep -w ${CONTAINER_NAME} > /dev/null; then

  # start Nominatim container exposing DB to the host
  docker run --rm -d --shm-size=8g \
    -e PBF_URL=https://download.geofabrik.de/europe/germany/berlin-latest.osm.pbf \
    -e REPLICATION_URL=https://download.geofabrik.de/europe/germany/berlin-updates/ \
    -e IMPORT_WIKIPEDIA=false \
    -e FREEZE=true \
    -v ${NOMINATIM_DATA}:/var/lib/postgresql/12/main \
    -p 8081:8080 \
    -p 5432:5432 \
    --name ${CONTAINER_NAME} \
    mediagis/nominatim:4.0
fi

# Wait for the Nominatim container to accept HTTP connections
# (this may a long time, if the container is started for the first time).
while [[ $(curl -s -o /dev/null -w ''%{http_code}'' "${CONTAINER_CHECK_URL}") != "200" ]]; do
  echo "waiting for ${CONTAINER_NAME} to accept connections"; sleep 5;
done


# get the DB password from the container
export PGPASSWORD="$(docker exec -it ${CONTAINER_NAME} /bin/bash -c 'echo -n $NOMINATIM_PASSWORD')"

# create a (tmp) table with desired data
#psql -h localhost -U nominatim -d nominatim -a -f ${SQL_FILE}

# dump the (tmp) table to a CSV file
psql -h localhost -U nominatim -d nominatim -c "\COPY districts_dump TO ${CSV_FILE_DISTRICTS} CSV HEADER;"
psql -h localhost -U nominatim -d nominatim -c "\COPY streets_dump TO ${CSV_FILE_STREETS} CSV HEADER;"
psql -h localhost -U nominatim -d nominatim -c "\COPY locations_dump TO ${CSV_FILE_LOCATIONS} CSV HEADER;"
psql -h localhost -U nominatim -d nominatim -c "\COPY housenumbers_dump TO ${CSV_FILE_HOUSENUMBERS} CSV HEADER;"

# try to stop the container
#docker stop nominatim || true
