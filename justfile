# just -l to list available commands

image := "price-tracker"
tag := "local"
go_migrate_version := "v4.18.2"
dbuser := "olxtracker"
dbpassword := "olxtracker"

build-local-docker:
    docker build -t {{image}}:{{tag}} .

# Go migrate for local development. This is generally only needed to create
# valid migration names (see new-migration).
install-go-migrate-linux:
    mkdir -p .bin
    curl -L --silent https://github.com/golang-migrate/migrate/releases/download/{{go_migrate_version}}/migrate.linux-amd64.tar.gz | tar xvz --directory .bin migrate
    chmod +x .bin/migrate

new-migration NAME:
	./.bin/migrate create -ext .sql -dir migrations {{NAME}}

docker-exec-db:
	docker compose exec postgres psql -U {{dbuser}} -d {{dbuser}}
