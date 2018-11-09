#!/usr/bin/env bash

function dumpMayaAPIServerLogs() {
  LC=$1
  MAPIPOD=$(kubectl get pods -o jsonpath='{.items[?(@.spec.containers[0].name=="maya-apiserver")].metadata.name}' -n openebs)
  kubectl logs --tail=${LC} $MAPIPOD -n openebs
  printf "\n\n"
}

echo "*****************************Deploying Openebs***************************"
CI_BRANCH="master"
CI_TAG="ci"

#Images from this repo are always tagged as ci
#The downloaded operator file will may contain a non-ci tag name
# depending on when and from where it is being downloaded. For ex:
# - during the release time, the image tags can be versioned like 0.7.0-RC..
# - from a branch, the image tags can be the branch names like v0.7.x-ci
if [ ${CI_TAG} != "ci" ]; then
  sudo docker tag openebs/openebs-k8s-provisioner:ci openebs/openebs-k8s-provisioner:${CI_TAG}
  sudo docker tag openebs/snapshot-controller:ci openebs/snapshot-controller:${CI_TAG}
  sudo docker tag openebs/snapshot-provisioner:ci openebs/snapshot-provisioner:${CI_TAG}
fi

#Tag the images with quay.io, since the operator can either have quay or docker images
sudo docker tag openebs/openebs-k8s-provisioner:ci quay.io/openebs/openebs-k8s-provisioner:${CI_TAG}
sudo docker tag openebs/snapshot-controller:ci quay.io/openebs/snapshot-controller:${CI_TAG}
sudo docker tag openebs/snapshot-provisioner:ci quay.io/openebs/snapshot-provisioner:${CI_TAG}

kubectl apply -f https://raw.githubusercontent.com/openebs/openebs/${CI_BRANCH}/k8s/openebs-operator.yaml

for i in $(seq 1 50) ; do
    replicas=$(kubectl get deployment -n openebs maya-apiserver -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
    else
        echo "Waiting Maya-apiserver to be ready"
        sleep 10
    fi
done

# wait for maya to complete installation
sleep 100

# Create deployment of snapshot controller & provisioner and RBAC policy for
# volumesnapshot API
echo "***********Deploying snapshot-controller and snapshot-provisioner********"
for i in $(seq 1 50) ; do
    replicas=$(kubectl get deployment -n openebs openebs-snapshot-operator -o json | jq ".status.readyReplicas")
    if [ "$replicas" == "1" ]; then
        break
	else
        echo "----------Snapshot deployment is not ready yet-------------------"
        sleep 10
    fi
done

# Install iscsi pkg
echo "Installing iscsi packages"
sudo apt-get update && sudo apt-get install open-iscsi
sudo service iscsid start
sudo service iscsid status

kubectl get pods --all-namespaces
kubectl get sc

echo "------------------------ Create sparse storagepoolclaim --------------- "
# delete the storagepoolclaim created earlier and create new spc with min/max pool
# count 1

for i in $(seq 1 30) ; do
    isDeleted=$(kubectl delete spc --all | grep deleted)
    if [ "$isDeleted" == "1" ]; then
        echo "Pool deleted successfully"
        break
    fi
    sleep 10
done

echo "------------------------ Create sparse storagepoolclaim --------------- "
# delete the storagepoolclaim created earlier and create new spc with min/max pool
# count 1
kubectl delete spc --all
kubectl apply -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/sample-pv-yamls/spc-sparse-single.yaml
sleep 10

echo "******************* Maya apiserver later logs *******************"
dumpMayaAPIServerLogs 100

echo "******************* Create Cstor and Jiva PersistentVolume *******************"
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/demo/pvc-single-replica-jiva.yaml
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/sample-pv-yamls/pvc-sparse-claim-cstor.yaml

dumpMayaAPIServerLogs 300

sleep 30

echo "******************* Describe disks **************************"
kubectl describe disks

echo "******************* Describe spc,sp,csp **************************"
kubectl describe spc,sp,csp

echo "******************* List all pods **************************"
kubectl get po --all-namespaces

echo "******************* List PVC,PV and pods **************************"
kubectl get pvc,pv

# Create the application
echo "Creating busybox-jiva and busybox-cstor application pod"
kubectl create -f $DST_REPO/external-storage/openebs/ci/snapshot/jiva/busybox.yaml
kubectl create -f $DST_REPO/external-storage/openebs/ci/snapshot/cstor/busybox.yaml

for i in $(seq 1 100) ; do
    phaseJiva=$(kubectl get pods busybox-jiva --output="jsonpath={.status.phase}")
    phaseCstor=$(kubectl get pods busybox-cstor --output="jsonpath={.status.phase}")
    if [ "$phaseJiva" == "Running" ] && [ "$phaseCstor" == "Running" ]; then
        break
	else
        echo "busybox-jiva pod is in:" $phaseJiva
        echo "busybox-cstor pod is in:" $phaseCstor

        if [ "$phaseJiva" != "Running" ]; then
           kubectl describe pods busybox-jiva
        fi
        if [ "$phaseCstor" != "Running" ]; then
           kubectl describe pods busybox-cstor
        fi
        sleep 10
    fi
done

dumpMayaAPIServerLogs 100

echo "********************Creating volume snapshot*****************************"
kubectl create -f  $DST_REPO/external-storage/openebs/ci/snapshot/jiva/snapshot.yaml
kubectl create -f  $DST_REPO/external-storage/openebs/ci/snapshot/cstor/snapshot.yaml
kubectl logs --tail=20 -n openebs deployment/openebs-snapshot-operator -c snapshot-controller

# It might take some time for cstor snapshot to get created. Wait for snapshot to get created
for i in $(seq 1 100) ; do
    kubectl get volumesnapshotdata
    count=$(kubectl get volumesnapshotdata | wc -l)
    # count should be 3 as one header line would also be present
    if [ "$count" == "3" ]; then
        break
    else
        echo "snapshot/(s) not created yet"
        kubectl get volumesnapshot,volumesnapshotdata
        sleep 10
    fi
done

kubectl logs --tail=20 -n openebs deployment/openebs-snapshot-operator -c snapshot-controller

# Promote/restore snapshot as persistent volume
sleep 30
echo "*****************Promoting snapshot as new PVC***************************"
kubectl create -f  $DST_REPO/external-storage/openebs/ci/snapshot/jiva/snapshot_claim.yaml
kubectl logs --tail=20 -n openebs deployment/openebs-snapshot-operator -c snapshot-provisioner
kubectl create -f  $DST_REPO/external-storage/openebs/ci/snapshot/cstor/snapshot_claim.yaml
kubectl logs --tail=20 -n openebs deployment/openebs-snapshot-operator -c snapshot-provisioner

sleep 30
# get clone replica pod IP to make a curl request to get the clone status
cloned_replica_ip=$(kubectl get pods -owide -l openebs.io/persistent-volume-claim=demo-snap-vol-claim-jiva --no-headers | grep -v ctrl | awk {'print $6'})
echo "***************** checking clone status *********************************"
for i in $(seq 1 5) ; do
		clonestatus=`curl http://$cloned_replica_ip:9502/v1/replicas/1 | jq '.clonestatus' | tr -d '"'`
		if [ "$clonestatus" == "completed" ]; then
            break
		else
            echo "Clone process in not completed ${clonestatus}"
            sleep 60
        fi
done

# Clone is in Alpha state, and kind of flaky sometimes, comment this integration test below for time being,
# util its stable in backend storage engine
echo "***************Creating busybox-clone-jiva application pod********************"
kubectl create -f $DST_REPO/external-storage/openebs/ci/snapshot/jiva/busybox_clone.yaml
kubectl create -f $DST_REPO/external-storage/openebs/ci/snapshot/cstor/busybox_clone.yaml

kubectl get pods --all-namespaces
kubectl get pvc --all-namespaces

for i in $(seq 1 15) ; do
    phaseJiva=$(kubectl get pods busybox-clone-jiva --output="jsonpath={.status.phase}")
    phaseCstor=$(kubectl get pods busybox-clone-cstor --output="jsonpath={.status.phase}")
    if [ "$phaseJiva" == "Running" ] && [ "$phaseCstor" == "Running" ]; then
        break
    else
        echo "busybox-clone-jiva pod is in:" $phaseJiva
        echo "busybox-clone-cstor pod is in:" $phaseCstor

        if [ "$phaseJiva" != "Running" ]; then
            kubectl describe pods busybox-clone-jiva
        fi
        if [ "$phaseCstor" != "Running" ]; then
            kubectl describe pods busybox-clone-cstor
        fi
		sleep 30
        fi
done


echo "********************** cvr status *************************"
kubectl get cvr -n openebs -o yaml

dumpMayaAPIServerLogs 100

kubectl get pods
kubectl get pvc

echo "*************Verifying data validity and Md5Sum Check********************"
hashjiva1=$(kubectl exec busybox-jiva -- md5sum /mnt/store1/date.txt | awk '{print $1}')
hashjiva2=$(kubectl exec busybox-clone-jiva -- md5sum /mnt/store2/date.txt | awk '{print $1}')

hashcstor1=$(kubectl exec busybox-cstor -- md5sum /mnt/store1/date.txt | awk '{print $1}')
hashcstor2=$(kubectl exec busybox-clone-cstor -- md5sum /mnt/store2/date.txt | awk '{print $1}')

echo "busybox jiva hash: $hashjiva1"
echo "busybox-clone-jiva hash: $hashjiva2"
echo "busybox cstor hash: $hashcstor1"
echo "busybox-clone-cstor hash: $hashcstor2"

if [ "$hashjiva1" != "" ] && [ "$hashcstor1" != "" ] && [ "$hashjiva1" == "$hashjiva2" ] && [ "$hashcstor1" == "$hashcstor2" ]; then
	echo "Md5Sum Check: PASSED"
else
    echo "Md5Sum Check: FAILED"; exit 1
fi
