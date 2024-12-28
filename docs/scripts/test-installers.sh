#!/bin/bash

# Script to test the installers on various docker images

set -eo pipefail
DOCKER_IMAGES=("ubuntu:latest" "frolvlad/alpine-bash" "gentoo/stage3:latest" "fedora:latest" "archlinux:latest" "opensuse/leap:latest" "debian:latest" "registry.access.redhat.com/ubi8/ubi:latest" "registry.suse.com/suse/sle15:latest" "oraclelinux:9" "amazonlinux:latest" "rockylinux/rockylinux:latest")

for image in "${DOCKER_IMAGES[@]}"
do
    echo "================== Testing $image =================="
    docker run --rm -t -v $(pwd)/docs/scripts:/install:ro -w /install $image /bin/bash /install/install.sh --yes
done