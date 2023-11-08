#! /bin/bash

helm upgrade --install zora undistro/zora \
  -n zora-system \
  --set scan.plugins.marvin.resources.limits.memory=256Mi  \
  --set scan.plugins.marvin.resources.limits.cpu=1Gi \
  --set clusterName=$(kubectl config current-context)