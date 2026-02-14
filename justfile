# just -l to list available commands

image := "price-tracker"
tag := "local"
go_migrate_version := "v4.18.2"
dbuser := "olxtracker"
dbpassword := "olxtracker"
test_compose_project := "testolx"
dev_compose_project := "devolx"

build-local-docker:
    docker build -t {{ image }}:{{ tag }} .

# Go migrate for local development. This is generally only needed to create

# valid migration names (see new-migration).
install-go-migrate-linux:
    mkdir -p .bin
    curl -L --silent https://github.com/golang-migrate/migrate/releases/download/{{ go_migrate_version }}/migrate.linux-amd64.tar.gz | tar xvz --directory .bin migrate
    chmod +x .bin/migrate

new-migration NAME:
    ./.bin/migrate create -ext .sql -dir migrations {{ NAME }}

docker-exec-db:
    docker compose -p {{ dev_compose_project }} exec postgres psql -U {{ dbuser }} -d {{ dbuser }}

dev-up:
    docker compose -p {{ dev_compose_project }} -f docker-compose.yaml up -d

dev-down:
    docker compose -p {{ dev_compose_project }} -f docker-compose.yaml down -v --remove-orphans

integration-test-local:
    #!/usr/bin/env bash
    set -euo pipefail

    cleanup() {
      docker compose -p {{ test_compose_project }} -f docker-compose.yaml -f docker-compose.test.yaml down -v --remove-orphans
    }

    trap cleanup EXIT

    docker compose -p {{ test_compose_project }} -f docker-compose.yaml -f docker-compose.test.yaml up -d
    RUN_INTEGRATION_TESTS=true OLXTRACKER_POSTGRES_PORT=5433 go test -count=1 -v ./...
