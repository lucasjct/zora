#!/bin/bash

function install_zora(){
  helm_commands=(
    "helm repo add undistro https://charts.undistro.io --force-update"
    "helm repo update undistro"
    "helm upgrade --install zora undistro/zora \
        -n zora-system \
        --version 0.7.0 \
        --create-namespace \
        --wait \
        --set clusterName=\"$(kubectl config current-context)\""
)

  for command in "${helm_commands[@]}"; do
      eval "$command"
  if [ $? -ne 0 ]; then
      echo "########## ERROR ##########"
      exit 1
  fi

done
}