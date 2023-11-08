#!/bin/bash

# check if schdule is corret witk 'grep' command
kubectl get cluster,scan -o wide -n zora-system
