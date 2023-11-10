

-- update sites_locations a,locations b set a.location_id=b.id where position(a.loc in b.name) and location_id=0;

-- select a.site,a.location as site_location,b.id master_location_id,b.name master_location_name ,s.id as surface_id,s.name as surface_name
-- from
--     sites_locations a left join
--     locations b on position(a.loc in b.name)
--     join surfaces s on s.location_id=b.id;


select a.site,a.location as site_location,a.surface as site_location_surface, b.id as location_master_id,b.name master_location,s.id as surface_id,s.name as surface_name
    from
        sites_locations a inner join
        locations b on a.location_id = b.id left join
        surfaces s on s.location_id=b.id and position(a.surface in s.name)<>0
        where a.site="omha-aaa" and s.id is not null and a.surface<>"" order by a.site, a.location ;

update sites_locations set loc= regexp_replace(location, ' \\(.+\\)','');
update sites_locations set surface=regexp_substr(location, '\\(.+\\)');
update sites_locations set surface=regexp_replace(surface, '\\(', '');
update sites_locations set surface=regexp_replace(surface, '\\)', '');
update sites_locations a, locations b, surfaces s set a.surface_id=s.id where a.location_id=b.id and s.location_id=b.id and position(a.surface in s.name)<>0 and s.id is not null and a.surface<>"";

--
select s.id,s.name, l.id, l.name from surfaces s join locations l on s.location_id=l.id where l.id in(select location_id from sites_locations where surface="")
============
update sites_locations set location_id = null, match_type = null;

update sites_locations s, locations l set s.location_id = l.id, s.match_type="postal code" where l.postal_code<>"" and position(l.postal_code in s.address);

update sites_locations s, locations l set s.location_id = l.id, s.match_type="partial" where position(regexp_substr(address1, '^[a-zA-Z0-9]+ [a-zA-Z0-9]+') in s.address) and position(left(l.postal_code,3) in s.address) and s.location_id is null;

update sites_locations s, locations l set s.location_id = l.id, s.match_type="address" where position(regexp_substr(address1, '^[a-zA-Z0-9]+ [a-zA-Z0-9]+') in s.address) and s.location_id is null;

update sites_locations s, locations l, surfaces r set s.surface_id=r.id where s.location_id=l.id and r.location_id=l.id and l.total_surfaces=1 and s.surface_id=0;
--update sites_locations a,locations b set a.location_id=b.id,match_type="name" where position(a.loc in b.name) and a.location_id is null and a.site="tcmhl";
=============
select * from sites_locations where location_id is null and site="tcmhl";
