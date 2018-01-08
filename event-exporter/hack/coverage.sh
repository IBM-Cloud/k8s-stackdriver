#!/bin/bash

echo 'mode: atomic' > cover.out
EXIT_CODE=0
for pkg in $(go list ./... | grep -v vendor); do
    go test -covermode=atomic -coverprofile=coverage.tmp $pkg || EXIT_CODE=$?
    if [ -f coverage.tmp ]; then
      tail -n +2 coverage.tmp >> cover.out
      rm coverage.tmp
    fi
done
go tool cover -html=cover.out -o=cover.html
exit $EXIT_CODE
