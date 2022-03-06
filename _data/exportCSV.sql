/*
 * districts
 */
drop table if exists districts;
create table districts (
	postcode varchar primary key,
	district varchar not null,
	geometry geometry(geometry,4326) not null,
	centroid geometry(geometry,4326) not null
);
with cte_district (postcode, district) as (
	values
	('10115', 'Mitte'),
	('10117', 'Mitte'),
	('10178', 'Mitte'),
	('10179', 'Mitte'),
	('10551', 'Mitte'),
	('10553', 'Mitte'),
	('10555', 'Mitte'),
	('10557', 'Mitte'),
	('10559', 'Mitte'),
	('10785', 'Mitte'),
	('13787', 'Mitte'),
	('13347', 'Mitte'),
	('13349', 'Mitte'),
	('13351', 'Mitte'),
	('13353', 'Mitte'),
	('13355', 'Mitte'),
	('13357', 'Mitte'),
	('13359', 'Mitte'),
	('13407', 'Mitte'),
	('10585', 'Charlottenburg-Wilmersdorf'),
	('10587', 'Charlottenburg-Wilmersdorf'),
	('10589', 'Charlottenburg-Wilmersdorf'),
	('10623', 'Charlottenburg-Wilmersdorf'),
	('10625', 'Charlottenburg-Wilmersdorf'),
	('10627', 'Charlottenburg-Wilmersdorf'),
	('10629', 'Charlottenburg-Wilmersdorf'),
	('10707', 'Charlottenburg-Wilmersdorf'),
	('10709', 'Charlottenburg-Wilmersdorf'),
	('10711', 'Charlottenburg-Wilmersdorf'),
	('10713', 'Charlottenburg-Wilmersdorf'),
	('10715', 'Charlottenburg-Wilmersdorf'),
	('10717', 'Charlottenburg-Wilmersdorf'),
	('10719', 'Charlottenburg-Wilmersdorf'),
	('13627', 'Charlottenburg-Wilmersdorf'),
	('14050', 'Charlottenburg-Wilmersdorf'),
	('14052', 'Charlottenburg-Wilmersdorf'),
	('14053', 'Charlottenburg-Wilmersdorf'),
	('14055', 'Charlottenburg-Wilmersdorf'),
	('14057', 'Charlottenburg-Wilmersdorf'),
	('14059', 'Charlottenburg-Wilmersdorf'),
	('14197', 'Charlottenburg-Wilmersdorf'),
	('14199', 'Charlottenburg-Wilmersdorf'),
	('10243', 'Friedrichshain-Kreuzberg'),
	('10245', 'Friedrichshain-Kreuzberg'),
	('10961', 'Friedrichshain-Kreuzberg'),
	('10963', 'Friedrichshain-Kreuzberg'),
	('10965', 'Friedrichshain-Kreuzberg'),
	('10967', 'Friedrichshain-Kreuzberg'),
	('10969', 'Friedrichshain-Kreuzberg'),
	('10997', 'Friedrichshain-Kreuzberg'),
	('10999', 'Friedrichshain-Kreuzberg'),
	('10315', 'Lichtenberg'),
	('10317', 'Lichtenberg'),
	('10318', 'Lichtenberg'),
	('10319', 'Lichtenberg'),
	('10365', 'Lichtenberg'),
	('10367', 'Lichtenberg'),
	('10369', 'Lichtenberg'),
	('10351', 'Lichtenberg'),
	('13053', 'Lichtenberg'),
	('13055', 'Lichtenberg'),
	('13057', 'Lichtenberg'),
	('13059', 'Lichtenberg'),
	('12619', 'Marzahn-Hellersdorf'),
	('12621', 'Marzahn-Hellersdorf'),
	('12623', 'Marzahn-Hellersdorf'),
	('12627', 'Marzahn-Hellersdorf'),
	('12629', 'Marzahn-Hellersdorf'),
	('12679', 'Marzahn-Hellersdorf'),
	('12681', 'Marzahn-Hellersdorf'),
	('12683', 'Marzahn-Hellersdorf'),
	('12685', 'Marzahn-Hellersdorf'),
	('12687', 'Marzahn-Hellersdorf'),
	('12689', 'Marzahn-Hellersdorf'),
	('12043', 'Neukölln'),
	('12045', 'Neukölln'),
	('12047', 'Neukölln'),
	('12049', 'Neukölln'),
	('12051', 'Neukölln'),
	('12053', 'Neukölln'),
	('12055', 'Neukölln'),
	('12057', 'Neukölln'),
	('12059', 'Neukölln'),
	('12347', 'Neukölln'),
	('12349', 'Neukölln'),
	('12351', 'Neukölln'),
	('12353', 'Neukölln'),
	('12355', 'Neukölln'),
	('12357', 'Neukölln'),
	('12359', 'Neukölln'),
	('10119', 'Mitte / Pankow'),
	('10247', 'Friedrichshain-Kreuzberg / Pankow'),
	('10249', 'Friedrichshain-Kreuzberg / Pankow'),
	('10405', 'Pankow'),
	('10407', 'Pankow'),
	('10409', 'Pankow'),
	('10435', 'Pankow'),
	('10437', 'Pankow'),
	('10439', 'Pankow'),
	('13051', 'Pankow'),
	('13086', 'Pankow'),
	('13088', 'Pankow'),
	('13089', 'Pankow'),
	('13125', 'Pankow'),
	('13127', 'Pankow'),
	('13129', 'Pankow'),
	('13156', 'Pankow'),
	('13158', 'Pankow'),
	('13159', 'Pankow'),
	('13187', 'Pankow'),
	('13189', 'Pankow'),
	('13047', 'Reinickendorf'),
	('13403', 'Reinickendorf'),
	('13405', 'Reinickendorf'),
	('13409', 'Mitte / Reinickendorf'),
	('13435', 'Reinickendorf'),
	('13437', 'Reinickendorf'),
	('13439', 'Reinickendorf'),
	('13465', 'Reinickendorf'),
	('13467', 'Reinickendorf'),
	('13469', 'Reinickendorf'),
	('13503', 'Reinickendorf'),
	('13505', 'Reinickendorf'),
	('13507', 'Reinickendorf'),
	('13509', 'Reinickendorf'),
	('13581', 'Spandau'),
	('13583', 'Spandau'),
	('13585', 'Spandau'),
	('13587', 'Spandau'),
	('13589', 'Spandau'),
	('13591', 'Spandau'),
	('13593', 'Spandau'),
	('13595', 'Spandau'),
	('13597', 'Spandau'),
	('13599', 'Spandau'),
	('13629', 'Spandau'),
	('14089', 'Spandau'),
	('12163', 'Steglitz-Zehlendorf'),
	('12165', 'Steglitz-Zehlendorf'),
	('12167', 'Steglitz-Zehlendorf'),
	('12203', 'Steglitz-Zehlendorf'),
	('12205', 'Steglitz-Zehlendorf'),
	('12207', 'Steglitz-Zehlendorf'),
	('12209', 'Steglitz-Zehlendorf'),
	('12247', 'Steglitz-Zehlendorf'),
	('12249', 'Steglitz-Zehlendorf'),
	('14109', 'Steglitz-Zehlendorf'),
	('14129', 'Steglitz-Zehlendorf'),
	('14163', 'Steglitz-Zehlendorf'),
	('14165', 'Steglitz-Zehlendorf'),
	('14167', 'Steglitz-Zehlendorf'),
	('14169', 'Steglitz-Zehlendorf'),
	('14193', 'Steglitz-Zehlendorf / Charlottenburg-Wilmersdorf'),
	('14195', 'Steglitz-Zehlendorf'),
	('10777', 'Tempelhof-Schöneberg'),
	('10779', 'Tempelhof-Schöneberg'),
	('10781', 'Tempelhof-Schöneberg'),
	('10783', 'Tempelhof-Schöneberg'),
	('10787', 'Tempelhof-Schöneberg'),
	('10789', 'Tempelhof-Schöneberg'),
	('10823', 'Tempelhof-Schöneberg'),
	('10825', 'Tempelhof-Schöneberg'),
	('10827', 'Tempelhof-Schöneberg'),
	('10829', 'Tempelhof-Schöneberg'),
	('12099', 'Tempelhof-Schöneberg'),
	('12101', 'Tempelhof-Schöneberg'),
	('12103', 'Tempelhof-Schöneberg'),
	('12105', 'Tempelhof-Schöneberg'),
	('12107', 'Tempelhof-Schöneberg'),
	('12109', 'Tempelhof-Schöneberg'),
	('12157', 'Steglitz-Zehlendorf / Tempelhof-Schöneberg'),
	('12159', 'Tempelhof-Schöneberg'),
	('12161', 'Steglitz-Zehlendorf / Tempelhof-Schöneberg'),
	('12169', 'Steglitz-Zehlendorf / Tempelhof-Schöneberg'),
	('12277', 'Tempelhof-Schöneberg'),
	('12279', 'Tempelhof-Schöneberg'),
	('12305', 'Tempelhof-Schöneberg'),
	('12307', 'Tempelhof-Schöneberg'),
	('12309', 'Tempelhof-Schöneberg'),
	('12435', 'Treptow-Köpenick'),
	('12437', 'Treptow-Köpenick'),
	('12439', 'Treptow-Köpenick'),
	('12459', 'Treptow-Köpenick'),
	('12487', 'Treptow-Köpenick'),
	('12489', 'Treptow-Köpenick'),
	('12524', 'Treptow-Köpenick'),
	('12526', 'Treptow-Köpenick'),
	('12555', 'Treptow-Köpenick / Marzahn-Hellersdorf'),
	('12557', 'Treptow-Köpenick'),
	('12527', 'Treptow-Köpenick'),
	('12559', 'Treptow-Köpenick'),
	('12587', 'Treptow-Köpenick'),
	('12589', 'Treptow-Köpenick')
)
insert into districts (postcode, district, geometry, centroid)
	select cte.postcode, cte.district, p.geometry, p.centroid  
	from cte_district cte
	left join placex p on p.postcode = cte.postcode 
	where p.class = 'boundary' and type = 'postal_code';
	 
/*
 * street names
 */
drop table if exists street_names;
create table street_names (
   id SERIAL primary key,
   name VARCHAR not null,
   unique (name)
);
insert into	street_names 
	(name)
    select
		distinct p.name -> 'name' as name
	from
		placex p
	where
			p.class = 'highway'
		and p.type in ('cycleway', 'footway', 'living_street', 'motorway_link', 'pedestrian', 'primary', 'primary_link', 'residential', 'secondary', 'secondary_link', 'tertiary', 'tertiary_link', 'trunk', 'trunk_link', 'unclassified')
		and p.name is not null
		and p.name->'name' != ''
;


/*
 * procedure for finding connected linestring (i.e. idententify different streets with the same name)
 * 
 * see: https://gis.stackexchange.com/a/94243
 */
drop procedure computeStreetLines;
CREATE OR REPLACE procedure computeStreetLines(integer) AS
$$
DECLARE
this_id bigint;
this_geom geometry;
cluster_id_match integer;

id_a bigint;
id_b bigint;

begin

-- create a table of lines for the given street id
drop table if exists lines;
create table lines as ( 
select
	p.place_id as id,
	p.geometry as geom 
from
	street_names s
left join placex p on p.name -> 'name' = s.name 
where
	(p.class = 'highway'
		and p.type in ('cycleway', 'footway', 'living_street', 'motorway_link', 'pedestrian', 'primary', 'primary_link', 'residential', 'secondary', 'secondary_link', 'tertiary', 'tertiary_link', 'trunk', 'trunk_link', 'unclassified')
			and p.name is not null
			and p.name->'name' != '')
	and s.id = $1
);
	
-- create clusters on lines
DROP TABLE IF EXISTS clusters;
CREATE TABLE clusters (cluster_id serial, ids bigint[], geom geometry);
CREATE INDEX ON clusters USING GIST(geom);

-- Iterate through linestrings, assigning each to a cluster (if there is an intersection)
-- or creating a new cluster (if there is not)
FOR this_id, this_geom IN SELECT id, geom FROM lines LOOP
  -- Look for an intersecting cluster.  (There may be more than one.)
  --SELECT cluster_id FROM clusters WHERE ST_Intersects(this_geom, clusters.geom)
  -- ST_Intersects fails, with bad data (e.g. Admiralstraße). Allow geometry to be 200m appart.
  -- there is likely no two different streets with the same name less than 500m appart.	
  SELECT cluster_id FROM clusters WHERE ST_DWithin(this_geom, clusters.geom, 500, false)
     LIMIT 1 INTO cluster_id_match;

  IF cluster_id_match IS NULL THEN
     -- Create a new cluster
     INSERT INTO clusters (ids, geom) VALUES (ARRAY[this_id], this_geom);
  ELSE
     -- Append line to existing cluster
     UPDATE clusters SET geom = ST_Union(this_geom, geom),
                          ids = array_prepend(this_id, ids)
      WHERE clusters.cluster_id = cluster_id_match;
  END IF;
END LOOP;


-- Iterate through the clusters, combining clusters that intersect each other
LOOP
    SELECT a.cluster_id, b.cluster_id FROM clusters a, clusters b 
     WHERE ST_Intersects(a.geom, b.geom)
       AND a.cluster_id < b.cluster_id
      INTO id_a, id_b;

    EXIT WHEN id_a IS NULL;
    -- Merge cluster A into cluster B
    UPDATE clusters a SET geom = ST_Union(a.geom, b.geom), ids = array_cat(a.ids, b.ids)
      FROM clusters b
     WHERE a.cluster_id = id_a AND b.cluster_id = id_b;

    -- Remove cluster B
    DELETE FROM clusters WHERE cluster_id = id_b;
END LOOP;
END;
$$ language plpgsql;

DROP TABLE IF EXISTS streets;
CREATE TABLE streets (
	id SERIAL primary key,
	name varchar,
	cluster_id integer, 
	place_ids bigint[], 
	geometry geometry,
	centroid geometry,
	postcode varchar,
	--unique(name, postcode),
	constraint fk_postcode
   		foreign key(postcode) 
   		 references districts(postcode)
	
);

/*
 * streets
 */
do $$ 
declare
    arow record;
  begin
    for arow in
      select * from street_names
      --select * from street_names where name = 'Ahornallee'
	  --select * from street_names limit 100
    loop
      RAISE NOTICE 'Calling computeStreetLines(%)', arow.id;
	  call computeStreetLines(arow.id);
 	  insert into streets (name, cluster_id, place_ids, geometry, centroid, postcode)
		select * from (
	 	  	select sl.*, d.postcode 
	 	  	from (		
	 	  		select arow.name, c.cluster_id, c.ids, c.geom, ST_ClosestPoint(c.geom, ST_Centroid(c.geom)) as centroid 
				from clusters c
			) as sl
			left join districts d on ST_Contains(d.geometry, sl.centroid)
		) as psl
		where psl.postcode is not null;
		commit;
    end loop;
  end;
$$;


-- function to match a given geometry to the closest street with the given name
drop function if exists matchStreet;
CREATE OR REPLACE function matchStreet(varchar, geometry) RETURNS integer as
$$
DECLARE street_id integer;
begin
	select dp.id into street_id from (
		select * from (
			select *, ST_Distance(s.geometry, $2) as distance
			from streets s 
			where s.name = $1
		) as d order by distance limit 1
	) as dp;
	RETURN street_id;
END;
$$ LANGUAGE plpgsql;


/*
 * housenumbers 
 */
drop table if exists housenumbers;
create table housenumbers (
   place_id integer,
   street_id integer,
   housenumber VARCHAR,
   postcode VARCHAR,
   centroid geometry,
   constraint fk_districts foreign key(postcode) references districts(postcode),
   constraint fk_street foreign key(street_id) references streets(id)
   --unique(street_id, housenumber)
);
insert into	housenumbers 
	(place_id, street_id, housenumber, postcode, centroid)
	select
		p.place_id,
        matchStreet(p.address -> 'street', p.centroid) as street_id,
		p.housenumber,
	    p.postcode,
	    p.centroid
	from
	    placex p
	where
	        p.postcode is not null 
	        and p.address is not null 
	        and p.address -> 'street' != ''
	        and p.housenumber is not null
	        and p.postcode in (select plz from bezirke)
			and p.housenumber ~ '^[0-9]+[a-z]*$'
	--limit 10000
;


/*
 * locations
 */
drop table if exists locations;
create table locations (
   place_id VARCHAR not null,
   street_id integer,
   class VARCHAR not null,
   type VARCHAR not null,
   name VARCHAR not null,
   housenumber VARCHAR,
   postcode VARCHAR,
   centroid geometry,
   constraint fk_districts foreign key(postcode) references districts(postcode),
   constraint fk_street foreign key(street_id) references streets(id)
);
insert into locations 
	(place_id, street_id, class, type, name, housenumber, postcode, centroid)
	select
		p.place_id,
		matchStreet(p.address -> 'street', p.centroid) as street_id,
	    p.class,
	    p.type,
	    p.name -> 'name' as name,
	    p.housenumber as housenumber,
	    p.postcode as postcode,
	    p.centroid
	from
	    placex p
	where
	        p.class in ('tourism', 'amenity')
	    and p.type is not null
	    and p.name is not null
	    and p.address is not null
	    and p.postcode is not null
	    and (p.class != 'tourism' or p.type in ('hotel', 'museum', 'hostel'))
	    and (p.class != 'amenity' or p.type in ('nightclub', 'pub', 'restaurant', 'house', 'cafe', 'biergarten', 'bar'))
	    and p.postcode in (select postcode from districts)
	    and p.name->'name' != ''
	    and p.address -> 'street' != '';


/*
 * tables ready for CSV export
 */ 
	   
drop table if exists districts_dump;
create table districts_dump as (
	select 
		postcode,
	 	district
	from districts
);	   

drop table if exists streets_dump;
create table streets_dump as (
	select 
	 	id,
	 	name, 
	 	cluster_id,
	 	postcode,
	 	ST_Y(centroid) as lat, 
  	 	ST_X(centroid) as lon
	from streets
);

drop table if exists housenumbers_dump;
create table housenumbers_dump as (
	select 
		street_id, 
		housenumber, 
		postcode,
		ST_Y(centroid) as lat, 
		ST_X(centroid) as lon
	from housenumbers
);

drop table if exists locations_dump;
create table locations_dump as (
	select 
		type,
	 	name,
		street_id,
		housenumber,
	 	postcode,
	 	ST_Y(centroid) as lat, 
  	 	ST_X(centroid) as lon
	from locations
	where street_id is not null
);
