#!/bin/bash

helm delete zora -n zora-system | grep -w "uninstalled"