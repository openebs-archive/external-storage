#!/bin/bash

# Copyright 2017 The OpenEBS Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DST_REPO="$GOPATH/src/github.com/kubernetes-incubator"
export DST_REPO

echo "Building openebs-provisioner"
export DIMAGE="openebs/openebs-k8s-provisioner"
cd $DST_REPO/external-storage/
make push-openebs-provisioner
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

echo "Building snapshot-controller and snapshot-provisioner"
cd $DST_REPO/external-storage/snapshot
export REGISTRY="openebs/"
export VERSION="ci"
make container
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

cd $DST_REPO/external-storage/
$DST_REPO/external-storage/openebs/ci/travis-ci.sh
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi
