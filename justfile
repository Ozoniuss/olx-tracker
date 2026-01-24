# justfile

image := "price-tracker"
tag := "local"

build-local-docker:
    docker build -t {{image}}:{{tag}} .
