#!/bin/bash

# use gprep to find zora-operator
kubectl get pods -n zora-system | grep -w "zora-operator"