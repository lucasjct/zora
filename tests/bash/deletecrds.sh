#! /bin/bash

kubectl get crd -o=name | grep --color=never 'zora.undistro.io' | xargs kubectl delete