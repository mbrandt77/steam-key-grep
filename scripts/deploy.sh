#!/usr/bin/env bash

set -e

CANONICAL_SCRIPT=$(readlink -e $0)
SCRIPT_DIR=$(dirname ${CANONICAL_SCRIPT})

print_usage(){
    cat <<EOM
        Usage:
        deploy.sh <tag>
EOM
exit 1
}

stop_old_container(){
    echo "stop old container"
    local container_id=$(podman ps -a | grep steam-key-grep | awk '{print $1}')
    podman stop $container_id || true
    podman rm $container_id || true
}

deploy_bot(){
    local image=$1
    echo "deploy steam-key-grep: ${image}"
    podman run --name steam-key-grep --restart on-failure -d "$image"
}

TAG=${1:-}
if [ -z "$TAG" ]
then
    print_usage
fi

stop_old_container

IMAGE="ghcr.io/mbrandt77/steam-key-grep:${TAG}"

deploy_bot "${IMAGE}"
