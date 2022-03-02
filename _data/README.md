# Getting DATA

[Nominatim](https://nominatim.org/) provides a container that is suitable for CSV extraction. The Nominatim container 
does all the heavy lifting and CSV data may be exported from the embedded PostgreSQL. To play around (manually), you 
may do:

1. start the container while exposing the PostgreSQL DB to the host
2. get the DB credentials from the container
3. connect to the DB using tooling on the host
4. work with the `placex`

For scripted export of CSV data see [`./extractCSV.sh`](./extractCSV.sh). 

## Manual

### Start the Container

Start a Nominatim Container with (e.g.) Berlin OSM data exposing the PostgreSQL DB you may run something like:

~~~~
docker run --rm --shm-size=8g \
    -e PBF_URL=https://download.geofabrik.de/europe/germany/berlin-latest.osm.pbf \
    -e REPLICATION_URL=https://download.geofabrik.de/europe/germany/berlin-updates/ \
    -e IMPORT_WIKIPEDIA=false \
    -e FREEZE=true \
    -v ${PWD}/nominatim-data:/var/lib/postgresql/12/main \
    -p 8081:8080 \
    -p 5432:5432 \
    --name nominatim \
    mediagis/nominatim:4.0
~~~~    

### Get the Credentials

Copy the password of the PostgreSQL-user `nominatim` from within the running Nominatim Container 
(env-variable `NOMINATIM_PASSWORD`) to the local environment variable `PGPASSWORD` via:

~~~~
export PGPASSWORD="`docker exec -it nominatim /bin/bash -c 'echo -n $NOMINATIM_PASSWORD'`"
~~~~

### Connect to the DB

As of now you may connect to the DB. 

If you want to connect from the CLI, make sure the PostgreSQL CLI tool `psql` is installed:

~~~~
sudo aptitude install postgresql-client
~~~~

and connect like:

~~~~
psql -h localhost -U nominatim nominatim
~~~~

### Work with `placex`

Within the Nominatim DB the `placex` table is of particular interest. Run something like:

~~~~
select 
    * 
from 
    placex p 
where 
        p.class = 'highway' 
    and p.type = 'living_street' 
    and p.name -> 'name' like 'Elisabeth-Feller%';
~~~~

## Scripted Export

The Bash script [`./extractCSV.sh`](./extractCSV.sh) automates the export of the data selected in
[`./extractCSV.sql`](./extractCSV.sql). Run the export like: 

~~~~
./extractCSV.sh
~~~~