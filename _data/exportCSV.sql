drop table if exists bezirke;

create table bezirke (
	plz varchar primary key,
	bezirk varchar not null
);

insert
	into
	bezirke (plz,
	bezirk)
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
('12589', 'Treptow-Köpenick');


/*
 * locactions 
 */
drop table if exists locations;

create table locations (
   id SERIAL primary key,
   class VARCHAR not null,
   type VARCHAR not null,
   name VARCHAR not null,
   street VARCHAR not null,
   housenumber VARCHAR,
   plz VARCHAR,
   lat float,
   lon float,
   constraint fk_bezirke
   		foreign key(plz) 
   		 references bezirke(plz)
);


insert into locations 
	(class, type, name, street, housenumber, plz, lat, lon)
	select
	    p.class,
	    p.type,
	    p.name -> 'name' as name,
	    p.address -> 'street' as street,
	    p.housenumber as housenumber,
	    p.postcode as plz,
	    ST_Y(p.centroid) as lat,
	    ST_X(p.centroid) as lon
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
	    and p.postcode in (select plz from bezirke)
	    and p.name->'name' != ''
	    and p.address -> 'street' != '';

/*
 * streets
 */
drop table if exists streets;

create table streets (
   id SERIAL primary key,
   name VARCHAR not null
);

insert into	streets 
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
 * buildings 
 */
drop table if exists buildings;

create table buildings (
   id SERIAL primary key,
   street_id integer,
   class VARCHAR not null,
   type VARCHAR not null,
   housenumber VARCHAR,
   plz VARCHAR,
   lat float,
   lon float,
   constraint fk_bezirke foreign key(plz) references bezirke(plz),
   constraint fk_street foreign key(street_id) references streets(id)
);

insert into	buildings 
	(street_id, class, type, housenumber, plz, lat, lon)
	select
		s.id,
	    p.class,
	    p.type,
	    p.housenumber,
	    p.postcode as plz,
	    ST_Y(p.centroid) as lat,
	    ST_X(p.centroid) as lon
	from
	    placex p
	join streets s on s.name = p.address -> 'street' 
	where
	        p.class in ('building')
	        and p.postcode is not null 
	        and p.housenumber is not null
	        and p.postcode in (select plz from bezirke)
			and p.housenumber ~ '^[0-9]+[a-z]*$'
;
