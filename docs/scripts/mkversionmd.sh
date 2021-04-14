#!/bin/bash

date=`date`
FAS_version=`cat ../../.version`

cat << EOF
---
# Copyright and Version
&copy; 2021 Hewlett Packard Enterprise Development LP



  FAS
${FAS_version};
${date}

EOF

