#!/bin/bash

helm repo add undistro https://charts.undistro.io --force-update
helm repo update undistro
helm upgrade --install zora undistro/zora \
  -n zora-system \
  --version 0.7.0 \
  --create-namespace \
  --wait \
  --set clusterName="$(kubectl config current-context)" \
  | grep -w "Visit our documentation for in-depth information: https://zora-docs.undistro.io"
