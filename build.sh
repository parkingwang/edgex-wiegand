#!/usr/bin/env bash

export DOCKER_CLI_EXPERIMENTAL=enabled

modules=( "endpoint" "trigger" )

makeModule() {
    for dir in ${modules[@]} ; do
        echo "###### Building module: ${dir}"
        cd ${dir}
        # Build linux dist
        GOOS=linux GOARCH=arm make -f ../Makefile $*
        GOOS=linux GOARCH=amd64 make -f ../Makefile $*
        GOOS=linux make -f ../Makefile manifest
        cd -
    done
}

makeModule image push
