#!/usr/bin/env bash

modules=( \
"endpoint" \
"trigger" \
)

makeModule() {
    for dir in ${modules[@]} ; do
        echo ">> Build module: ${dir}"
        cd ${dir}
        OSARCH=arm ./make.sh $*
        cd -
    done
}

makeModule $*
