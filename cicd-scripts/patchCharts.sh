#!/bin/bash
# Copyright (c) 2020 Red Hat, Inc.

# build image with charts and replace image in cluster
# Use current namespace
REPOPOD=$(kubectl get pods -o=name | grep multiclusterhub-repo | sed "s/^.\{4\}//")
NAMESPACE=$(echo $(oc project) | cut -d " " -f 3)
NAMESPACE=${NAMESPACE//\"}

echo ${NAMESPACE}
echo ${REPOPOD}
echo $PWD

#create a temp dir of charts, move only tgz, rm tmp folder
kubectl cp $PWD/multiclusterhub/charts ${NAMESPACE}/${REPOPOD}:tmp/
kubectl exec ${REPOPOD} -- sh -c 'rm -rf multiclusterhub/charts/*.tgz && mv tmp/*.tgz multiclusterhub/charts && rm -rf tmp'