# Copyright Contributors to the Open Cluster Management project

#!/bin/bash
# Copyright (c) 2020 Red Hat, Inc.

echo "Building test-image"

docker build . -f integration_tests/build/Dockerfile -t $1
