#!/bin/bash

SLEEP_STEP=1

function run {

    set -x;

    $@

    set +x;

    sleep $SLEEP_STEP;
}

run "sudo make clean-all"

run "docker-compose -f docker-compose-cli.yaml down --volumes --remove-orphans"

run "docker system prune --volumes -f"

docker images -a | grep "dev-peer" | awk '{print $3}' | xargs docker rmi