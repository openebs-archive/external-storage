#!/usr/bin/env bash

echo "*****************************Deploying Openebs***************************"
CI_BRANCH="v0.7.x"
CI_TAG="v0.7.x-ci"

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
sudo apt-get install open-iscsi
sudo service iscsid start
sudo service iscsid status

kubectl get pods --all-namespaces
kubectl get sc


echo "**********Deploy CAS templates configuration for Maya-apiserver**********"
#kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-pre-release-features.yaml

sleep 30
echo "**********Create Persistentvolumeclaim with single replica****************"
kubectl create -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/demo/pvc-single-replica-jiva.yaml

sleep 30
echo "******************* List PVC,PV and pods **************************"
kubectl get pvc,pv

# Create the application
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

echo "********************Creating volume snapshot*****************************"
kubectl create -f  $DST_REPO/external-storage/openebs/ci/snapshot/snapshot.yaml
kubectl logs --tail=20 -n openebs deployment/openebs-snapshot-operator -c snapshot-controller

# Promote/restore snapshot as persistent volume
sleep 30
echo "*****************Promoting snapshot as new PVC***************************"
kubectl create -f  $DST_REPO/external-storage/openebs/ci/snapshot/snapshot_claim.yaml
kubectl logs --tail=20 -n openebs deployment/openebs-snapshot-operator -c snapshot-provisioner

sleep 30
# get clone replica pod IP to make a curl request to get hte clone status
cloned_replica_ip=$(kubectl get pods -owide -l pvc=demo-snap-vol-claim --no-headers | grep -v ctrl | awk {'print $6'})
echo "***************** checking clone status *********************************"
for i in $(seq 1 100) ; do
		clonestatus=`curl http://$cloned_replica_ip:9502/v1/replicas/1 | jq '.clonestatus' | tr -d '"'`
		if [ "$clonestatus" == "completed" ]; then
        break
			else
        echo "Clone process in not completed ${clonestatus}"
        sleep 10
    fi
done

# Clone is in Alpha state, and kind of flaky sometimes, comment this integration test below for time being,
# util its stable in backend storage engine
#sleep 30
#echo "***************Creating busybox-clone application pod********************"
#kubectl create -f $DST_REPO/external-storage/openebs/ci/snapshot/busybox_clone.yaml
#
#kubectl get pods --all-namespaces
#kubectl get pvc --all-namespaces
#
#for i in $(seq 1 50) ; do
#    phase=$(kubectl get pods busybox-clone --output="jsonpath={.status.phase}")
#    if [ "$phase" == "Running" ]; then
#        break
#    else
#        echo "--------------busybox-clone pod is not ready yet-----------------"
#        kubectl describe pods busybox-clone
#    sleep 10
#    fi
#done
#
#kubectl get pods
kubectl get pvc
#
#echo "*************Varifying data validity and Md5Sum Check********************"
#hash1=$(kubectl exec busybox -- md5sum /mnt/store1/date.txt | awk '{print $1}')
#hash2=$(kubectl exec busybox-clone -- md5sum /mnt/store2/date.txt | awk '{print $1}')
#echo "busybox hash: $hash1"
#echo "busybox-clone hash: $hash2"
#if [[ $hash1 == $hash2 ]]; then
#	 echo "Md5Sum Check: PASSED"
#else
#echo "Md5Sum Check: FAILED"; exit 1
#fi
