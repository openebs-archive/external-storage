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

#This is a forked repo. Copy the contents to the expected go folder.
SRC_REPO=`pwd`
DST_REPO="$GOPATH/src/github.com/kubernetes-incubator"
mkdir -p $DST_REPO
cp -R $SRC_REPO/../external-storage $DST_REPO

echo "Building openebs-provisioner"
cd $DST_REPO/external-storage/openebs
make container
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

echo "Building snapshot-controller and snapshot-provisioner"
cd $DST_REPO/external-storage/snapshot
export VERSION="ci"
make container
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi
