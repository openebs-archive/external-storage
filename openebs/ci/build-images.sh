#!/usr/bin/env bash

echo "*****************************Retagging images and setting up env***************************"

#Images from this repo are always tagged as ci
#The downloaded operator file will may contain a non-ci tag name
# depending on when and from where it is being downloaded. For ex:
# - during the release time, the image tags can be versioned like 0.7.0-RC..
# - from a branch, the image tags can be the branch names like v0.7.x-ci
if [ ${CI_TAG} != "ci" ]; then
  #sudo docker tag openebs/openebs-k8s-provisioner:ci openebs/openebs-k8s-provisioner:${CI_TAG}
  sudo docker tag openebs/snapshot-controller:ci openebs/snapshot-controller:${CI_TAG}
  sudo docker tag openebs/snapshot-provisioner:ci openebs/snapshot-provisioner:${CI_TAG}
fi

#Tag the images with quay.io, since the operator can either have quay or docker images
#sudo docker tag openebs/openebs-k8s-provisioner:ci quay.io/openebs/openebs-k8s-provisioner:${CI_TAG}
sudo docker tag openebs/snapshot-controller:ci quay.io/openebs/snapshot-controller:${CI_TAG}
sudo docker tag openebs/snapshot-provisioner:ci quay.io/openebs/snapshot-provisioner:${CI_TAG}

# Install iscsi pkg
echo "Installing iscsi packages"
sudo apt-get update && sudo apt-get install open-iscsi
sudo service iscsid start
sudo service iscsid status
echo "Installation complete"
