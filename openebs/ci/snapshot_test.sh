#!/usr/bin/env bash

# Create deployment of snapshot controller & provisioner and RBAC policy for
# volumesnapshot API
echo "Deploying snapshot-controller and snapshot-provisioner"
kubectl create -f snapshot/snapshot-operator.yaml

# Creates a snapshot
sleep 30
kubectl get pods

for i in $(seq 1 100) ; do
    replicas=$(kubectl get deployment snapshot-controller -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
    else
        echo "deployment is not ready yet"
        sleep 10
    fi
done
echo "Creating snapshot"
kubectl create -f snapshot/snapshot.yaml

# Created snapshot promoter storage-class
echo "Creating snapshot-promoter storage class"
kubectl create -f snapshot/snapshot_sc.yaml

# Promote/restore snapshot as persistent volume
sleep 60
echo "Promoting snapshot as new PVC"
kubectl create -f snapshot/snapshot_claim.yaml
