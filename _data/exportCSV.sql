drop table if exists dump;

create table dump as
    select
        p.place_id,
        p.parent_place_id,
        p.class,
        p.type,
        p.name -> 'name' as name,
        p.address -> 'street' as street,
        p.housenumber as house_number,
        p.address -> 'suburb' as suburb,
        p.postcode,
        p.address -> 'city' as city,
        ST_Y(p.centroid) as lat,
        ST_X(p.centroid) as lon
    from
        placex p
    where
            p.class in ('tourism', 'amenity', 'highway')
        and p.name is not null
        and p.address is not null
        and p.postcode is not null
        and p.parent_place_id != 0 -- drop apperently brocken entries
        and (class != 'highway' or type in ('cycleway', 'footway', 'living_street', 'motorway_link', 'pedestrian', 'primary', 'primary_link', 'residential', 'secondary', 'secondary_link', 'tertiary', 'tertiary_link', 'trunk', 'trunk_link', 'unclassified'))
        and (class != 'tourism' or type in ('hotel', 'museum', 'hostel'))
        and (class != 'amenity' or type in ('nightclub', 'pub', 'restaurant', 'house', 'cafe', 'biergarten', 'bar'))
        and p.name->'name' != ''
        and (p.class = 'highway' or (p.address -> 'street' is not null and p.address -> 'city' is not null and p.housenumber is not null));
;


-- drop duplicate entries for streets with the same name and the same zip
delete from dump d where d.class='highway' and d.place_id not in (select DISTINCT ON (d.name, d.postcode) d.place_id from dump d where d.class='highway');

