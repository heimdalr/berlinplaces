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
        case 
        	-- it's a street and the direct parent is a admin boundry take this
        	when 	p.class = 'highway' 
        		and pp.class = 'boundary' 
        		and pp.type = 'administrative' 
        		and pp.name is not null 
        		and pp.name -> 'name' != '' 
        	then pp.name -> 'name'
        	-- it's a street, the parent is a neighbourhood and the grand parent is a admin boundry take this
        	when 	p.class = 'highway' 
        		and pp.class = 'place' 
        		and pp.type = 'neighbourhood' 
        		and ppp.class = 'boundary' 
        		and ppp.type = 'administrative' 
        		and ppp.name is not null 
        		and ppp.name -> 'name' != '' 
        	then ppp.name -> 'name'  
        	-- if neither the direct parent nor the grand parent is a boundry ignore it
        	else ''
        end as boundary,
        case 
        	when 	p.class = 'highway' 
        		and pp.class = 'place' 
        		and pp.type = 'neighbourhood' 
        		and pp.name is not null 
        		and pp.name -> 'name' != '' 
        	then pp.name -> 'name'  
        	else ''
        end as neighbourhood,
        p.address -> 'suburb' as suburb,
        p.postcode,
        p.address -> 'city' as city,
        ST_Y(p.centroid) as lat,
        ST_X(p.centroid) as lon
    from
        placex p
    left join placex pp on pp.place_id = p.parent_place_id -- parent 
    left join placex ppp on ppp.place_id = pp.parent_place_id -- grandparent
    where
            p.class in ('tourism', 'amenity', 'highway')
        and p.name is not null
        and p.address is not null
        and p.postcode is not null
        and p.parent_place_id != 0 -- drop apperently brocken entries
        and (p.class != 'highway' or p.type in ('cycleway', 'footway', 'living_street', 'motorway_link', 'pedestrian', 'primary', 'primary_link', 'residential', 'secondary', 'secondary_link', 'tertiary', 'tertiary_link', 'trunk', 'trunk_link', 'unclassified'))
        and (p.class != 'tourism' or p.type in ('hotel', 'museum', 'hostel'))
        and (p.class != 'amenity' or p.type in ('nightclub', 'pub', 'restaurant', 'house', 'cafe', 'biergarten', 'bar'))
        and p.name->'name' != ''
        and (p.class = 'highway' or (p.address -> 'street' is not null and p.address -> 'city' is not null and p.housenumber is not null));
;


-- drop duplicate entries for streets with the same name and the same zip
delete from dump d where d.class='highway' and d.place_id not in (select DISTINCT ON (d.name, d.postcode) d.place_id from dump d where d.class='highway');

