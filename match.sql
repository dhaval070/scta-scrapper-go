update sites_locations set loc = null where loc = "" and site="tcmhl";
update sites_locations set surface = null where surface = "";

update sites_locations set loc= regexp_replace(location, ' \\(.+\\)','') where loc is null;
update sites_locations set surface=regexp_substr(location, '\\(.+\\)') where surface is null;
update sites_locations set surface=regexp_replace(surface, '\\(', '');
update sites_locations set surface=regexp_replace(surface, '\\)', '');


update sites_locations a,locations b set a.location_id=b.id where position(a.loc in b.name) and location_id=0;

select a.site,a.location as site_location,b.id master_location_id,b.name master_location_name ,s.id as surface_id,s.name as surface_name
from
    sites_locations a left join
    locations b on position(a.loc in b.name)
    join surfaces s on s.location_id=b.id;



select a.site,a.location as site_location,b.id,b.name master_location,s.id as surface_id,s.name as surface_name
    from
        sites_locations a left join
        locations b on position(a.loc in b.name)
        left join surfaces s on s.location_id=b.id;


update sites_locations set loc_id_by_zip=0;
update sites_locations a, locations b set a.loc_id_by_zip=b.id where position(replace(b.postal_code, " ","") in replace(a.address," ","")) and b.postal_code<>"";

============
update sites_locations set loc_id_1 = null, match_type = null where site="tcmhl";

update sites_locations s, locations l set s.loc_id_1 = l.id, s.match_type="postal code" where l.postal_code<>"" and position(left(l.postal_code, 3) in s.address) and s.site="tcmhl";


update sites_locations s, locations l set s.loc_id_1 = l.id, s.match_type="address" where 
position(regexp_substr(address1, '^[a-zA-Z0-9]+ [a-zA-Z0-9]+') in s.address) and s.loc_id_1 is null and s.site="tcmhl";

update sites_locations a,locations b set a.loc_id_1=b.id,match_type="name" where position(a.loc in b.name) and a.loc_id_1 is null and a.site="tcmhl";
=============
select * from sites_locations where loc_id_1 is null and site="tcmhl";
