update sites_locations set loc= regexp_replace(location, ' \\(.+\\)','');
update sites_locations set surface=regexp_substr(location, '\\(.+\\)');
update sites_locations set surface=regexp_replace(surface, '\\(', '');


update sites_locations a,locations b set a.location_id=b.id where position(a.loc in b.name);
