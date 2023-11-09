#! /bin/bash


succeslist="Thank you for installing Zora version 0.7.0."
succesnotes="zora-system"

echo "merda"

if helm list -n zora-system | grep -w $succeslist;
then
echo "matched"
fi