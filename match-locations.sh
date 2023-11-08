#!/bin/bash
site="$1"

if [ -z "$site" ]; then
    echo "site parameter is required"
fi

qry="
update sites_locations set location_id = null, match_type = null where site='$site';

update sites_locations s, locations l set s.location_id = l.id, s.match_type='postal code' where l.postal_code<>'' and position(left(l.postal_code, 3) in s.address) and s.site='$site';


update sites_locations s, locations l set s.location_id = l.id, s.match_type='address' where
position(regexp_substr(address1, '^[a-zA-Z0-9]+ [a-zA-Z0-9]+') in s.address) and s.location_id is null and s.site='$site';

update sites_locations a,locations b set a.location_id=b.id,match_type='name' where position(a.loc in b.name) and a.location_id is null and a.site='$site';"

echo $qry
