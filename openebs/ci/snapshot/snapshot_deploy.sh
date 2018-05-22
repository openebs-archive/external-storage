#!/usr/bin/env bash

echo "Deploying Openebs"
kubectl create -f $DST_REPO/external-storage/openebs/ci/snapshot/openebs-operator.yaml

for i in $(seq 1 50) ; do
    replicas=$(kubectl get deployment maya-apiserver -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
			else
        echo "Waiting Maya-apiserver to be ready"
        sleep 10
    fi
done

# Create deployment of snapshot controller & provisioner and RBAC policy for
# volumesnapshot API
echo "Deploying snapshot-controller and snapshot-provisioner"
kubectl create -f $DST_REPO/external-storage/openebs/ci/snapshot/snapshot-operator.yaml

for i in $(seq 1 50) ; do
    replicas=$(kubectl get deployment snapshot-controller -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
			else
        echo "Snapshot deployment is not ready yet"
        sleep 10
    fi
done

# Install iscsi pkg
echo "Installing iscsi packages"
sudo apt-get install open-iscsi
sudo service iscsid start
sudo service iscsid status

# Creates a snapshot
kubectl get pods --all-namespaces
echo "Creating busybox application pod"
kubectl create -f $DST_REPO/external-storage/openebs/ci/snapshot/busybox.yaml

for i in $(seq 1 100) ; do
    phase=$(kubectl get pods busybox --output="jsonpath={.status.phase}")
    if [ "$phase" == "Running" ]; then
        break
			else
        echo "busybox pod is not ready yet"
        kubectl describe pods busybox
        sleep 10
    fi
done

echo "Creating snapshot"
kubectl create -f  $DST_REPO/external-storage/openebs/ci/snapshot/snapshot.yaml
kubectl logs --tail=20 deployment/snapshot-controller -c snapshot-controller

# Promote/restore snapshot as persistent volume
sleep 30
echo "Promoting snapshot as new PVC"
kubectl create -f  $DST_REPO/external-storage/openebs/ci/snapshot/snapshot_claim.yaml
kubectl logs --tail=20 deployment/snapshot-controller -c snapshot-provisioner

sleep 30
echo "Creating busybox-clone application pod"
kubectl create -f $DST_REPO/external-storage/openebs/ci/snapshot/busybox_clone.yaml

kubectl get pods --all-namespaces
kubectl get pvc --all-namespaces

for i in $(seq 1 50) ; do
    phase=$(kubectl get pods busybox-clone --output="jsonpath={.status.phase}")
    if [ "$phase" == "Running" ]; then
        break
    else
        echo "busybox-clone pod is not ready yet"
        kubectl describe pods busybox-clone
    sleep 10
    fi
done

kubectl get pods
kubectl get pvc

echo "Varifying data validity and Md5Sum Check"
hash1=$(kubectl exec busybox -- md5sum /mnt/store1/date.txt | awk '{print $1}')
hash2=$(kubectl exec busybox-clone -- md5sum /mnt/store2/date.txt | awk '{print $1}')
echo "busybox hash: $hash1"
echo "busybox-clone hash: $hash2"
if [[ $hash1 == $hash2 ]]; then
	 echo "Md5Sum Check: PASSED"
else
echo "Md5Sum Check: FAILED"; exit 1
fi
