#!/usr/bin/env bash

echo "mode: atomic" > coverage.txt

failed=0
failed_packages=()

for d in $(go list ./... | grep -v vendor); do
    if ! go test -v -race -coverprofile=profile.out -covermode=atomic $d; then
        failed=1
        failed_packages+=("$d")
    fi
    if [ -f profile.out ]; then
        grep -v "^mode:" profile.out >> coverage.txt
        rm profile.out
    fi
done

echo
if [ $failed -eq 0 ]; then
    echo "TESTS PASSED: all packages returned 0"
else
    echo "TESTS FAILED in the following packages:"
    for p in "${failed_packages[@]}"; do
        echo "  - $p"
    done
fi

exit $failed
