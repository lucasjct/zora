#! /bin/bash

helm upgrade --install zora undistro/zora \
  -n zora-system \
  --set scan.misconfiguration.schedule="0 * * * *" \
  --set scan.vulnerability.schedule="0 0 * * *" \
  --set clusterName=$(kubectl config current-context)