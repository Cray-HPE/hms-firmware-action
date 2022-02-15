#! /usr/bin/env bash

if [[ "$1" == "smoke" ]]; then
    echo "Running smoke tests..."
    /opt/cray/tests/ncn-smoke/hms/hms-firmware-action/fas_smoke_test_ncn-smoke.sh
elif [[ "$1" == "functional" ]]; then
    echo "Running functional tests..."  
    /opt/cray/tests/ncn-functional/hms/hms-firmware-action/fas_tavern_api_test_ncn-functional.sh
else
    echo "Unsuported test type"
fi