#mysql --port=5506 -uapp -h172.17.0.1 -papp schedules "delete from sites_locations where site='$1'"
go run ./cmd/sites/$1/main.go -outfile var/20231108/$1.csv --import-locations
