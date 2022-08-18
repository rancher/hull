#!/bin/bash
set -e

go_test_tags="$@"

for module in $(find . -name 'go.mod' | sed 's/\/go.mod//'); do
    pushd ${module} 2>/dev/null 1>/dev/null
    go mod download

    popd 2>/dev/null 1>/dev/null
done

root_dir=$(pwd)
for module in $(find . -name 'go.mod' | sed 's/\/go.mod//'); do
    pushd ${module} 2>/dev/null 1>/dev/null
    if [[ -z ${PLUGIN_MODE} ]]; then
        go test -tags="${go_test_tags}" -v ./...
    else
        go test -tags="${go_test_tags}" -v 2>&1 ./... | tee -a ${root_dir}/go_test.log
    fi
    popd 2>/dev/null 1>/dev/null
done

if [[ -n ${PLUGIN_MODE} ]]; then
    # Ensures output matches Sonobuoy plugin expectations
    # ref: https://sonobuoy.io/docs/v0.51.0/plugins/
    go-junit-report -set-exit-code -in go_test.log > report.xml
    mkdir -p /tmp/results
    echo "$(pwd)/report.xml" > /tmp/results/done
fi