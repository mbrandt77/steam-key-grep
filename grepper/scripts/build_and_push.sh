#!/usr/bin/env bash

set -e

CANONICAL_SCRIPT=$(readlink -e $0)
SCRIPT_DIR=$(dirname ${CANONICAL_SCRIPT})
ROOT_DIR=$(realpath "${SCRIPT_DIR}/../")
BUILD_DIR=$(realpath "${ROOT_DIR}/build/")

print_usage(){
    cat <<EOM
        Usage:
        build_and_push.sh
EOM
exit 1
}

build_image(){
    local image=$1

    echo "build image: $image"
    buildah bud -t "${image}" -f "${ROOT_DIR}/Dockerfile" ${ROOT_DIR} 
  }

push_image(){
    local image=$1

    echo "push image: ${image}"
    buildah push $image
}

purge_none_images(){
    buildah rmi -f $(buildah images -a | grep '<none>' | awk '{print $3}') || true

    podman rmi -f $(podman images -a | grep '<none>' | awk '{print $3}') || true

}


GIT_SHORT_SHA=$(git rev-parse --short HEAD)
IMAGE="ghcr.io/mbrandt77/steam-key-grep:${GIT_SHORT_SHA}"

build_image $IMAGE
push_image $IMAGE

purge_none_images
echo "${GIT_SHORT_SHA}"