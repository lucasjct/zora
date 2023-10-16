#!/bin/bash

# Install Zora through Helm 
helm repo update undistro
helm upgrade --install zora undistro/zora \
  -n zora-system \
  --create-namespace \
  --wait \
  --set 'scan.plugins.trivy.ignoreDescriptions=true' \
  --set 'scan.vulnerability.schedule=* 3 * * *' \
  --set clusterName=$(kubectl config current-context) \
  --version 0.7.0-rc6
