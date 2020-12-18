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

# Determine the arch/os we're building for
ARCH=$(uname -m)

if [ -z "${IMAGE_ORG}" ]; then
	IMAGE_ORG="openebs"
fi

if [ "${ARCH}" = "x86_64" ]; then
	#export DIMAGE="${IMAGE_ORG}/openebs-k8s-provisioner"
	#./openebs/buildscripts/push

	export DIMAGE="${IMAGE_ORG}/snapshot-controller"
	./openebs/buildscripts/push

	export DIMAGE="${IMAGE_ORG}/snapshot-provisioner"
	./openebs/buildscripts/push
elif [ "${ARCH}" = "aarch64" ]; then
	#export DIMAGE="${IMAGE_ORG}/openebs-k8s-provisioner-arm64"
	#./openebs/buildscripts/push

	export DIMAGE="${IMAGE_ORG}/snapshot-controller-arm64"
	./openebs/buildscripts/push

	export DIMAGE="${IMAGE_ORG}/snapshot-provisioner-arm64"
	./openebs/buildscripts/push
else
	echo "${ARCH} is not supported"
	exit 1
fi


