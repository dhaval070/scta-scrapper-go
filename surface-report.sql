SELECT e.surface_id, l.name location_name, s.name surface_name, date_format(e.datetime, "%W") dow, min(e.datetime) start_time, max( date_add(e.datetime, INTERVAL 150 minute)) end_time
FROM
`events` e JOIN surfaces s on e.surface_id=s.id JOIN locations l on l.id=s.location_id
GROUP BY location_name, surface_name, surface_id, date(e.datetime)
ORDER BY location_name, surface_name, surface_id,dayofweek(e.datetime);
