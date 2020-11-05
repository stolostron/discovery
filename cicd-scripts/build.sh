#!/bin/bash
# Copyright (c) 2020 Red Hat, Inc.

echo "<repo>/<component>:<tag> : $1"

docker build . -t $1