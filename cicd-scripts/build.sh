# Copyright Contributors to the Open Cluster Management project

#!/bin/bash
# Copyright (c) 2020 Red Hat, Inc.

echo "<repo>/<component>:<tag> : $1"

docker build . -t $1
