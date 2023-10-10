update sites_locations set loc= regexp_replace(location, ' \\(.+\\)','');
update sites_locations set surface=regexp_substr(location, '\\(.+\\)');
update sites_locations set surface=regexp_replace(surface, '\\(', '');


update sites_locations a,locations b set a.location_id=b.id where position(a.loc in b.name);

select a.site,a.location as site_location,b.id,b.name master_location,s.id as surface_id,s.name as surface_name
from
    sites_locations a left join
    locations b on position(a.loc in b.name)
    join surfaces s on s.location_id=b.id;



select a.site,a.location as site_location,b.id,b.name master_location,s.id as surface_id,s.name as surface_name
    from
        sites_locations a left join
        locations b on position(a.loc in b.name)
        left join surfaces s on s.location_id=b.id;
