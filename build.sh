#!/usr/bin/env bash

export DOCKER_CLI_EXPERIMENTAL=enabled

SUDO="sudo "
if [ "Darwin" == "$(uname -s)" ]; then
    SUDO=""
fi

modules=( "node" )

makeModule() {
    for dir in ${modules[@]} ; do
        echo "###### Building module: ${dir}"
        cd ${dir}
        # Build linux dist
        OS_SUDO=${SUDO} GOOS=linux GOARCH=arm make -f ../Makefile $*
        OS_SUDO=${SUDO} GOOS=linux GOARCH=arm64 make -f ../Makefile $*
        OS_SUDO=${SUDO} GOOS=linux GOARCH=amd64 make -f ../Makefile $*
        OS_SUDO=${SUDO} GOOS=linux make -f ../Makefile manifest
        cd -
    done
}

makeModule image push
