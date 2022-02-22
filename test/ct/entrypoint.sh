#! /usr/bin/env bash

if [[ "$1" == "smoke" ]]; then
    echo "Running smoke tests..."
    /src/app/smoke_test.sh
elif [[ "$1" == "functional" ]]; then
    echo "Running functional tests..."  
    /src/app/functional_test.sh
else
    echo "Unsuported test type"
fi