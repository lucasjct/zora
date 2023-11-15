#! /bin/bash


# Function for edit resources
function edit_resource(){
  helm_resources=(
  "helm upgrade --install zora undistro/zora \
  -n zora-system \
  --set scan.plugins.marvin.resources.limits.memory=256Mi \
  --set scan.plugins.marvin.resources.limits.cpu=1Gi \
  --set clusterName=\"$(kubectl config current-context)\""
  )


for command in "${helm_resources[@]}"; do 
    eval "$command"
if [ $? -ne 0 ]; then 
    echo "########## ERROR ##########"
    exit 1
fi
done
}


# Function for edit schedule
function edit_schedule(){
  helm_schedule=(
    "helm upgrade --install zora undistro/zora \
    -n zora-system \
    --set scan.misconfiguration.schedule='0 * * * *' \
    --set scan.vulnerability.schedule='0 0 * * *' \
    --set clusterName=\"$(kubectl config current-context)\""
  )
for command in "${helm_schedule[@]}"; do 
    eval "$command"
if [ $? -ne 0 ]; then 
    echo "########## ERROR ##########"
    exit 1
fi
done
}