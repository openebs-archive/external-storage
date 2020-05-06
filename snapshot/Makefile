# Copyright 2017 The Kubernetes Authors.
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

# set the shell to bash in case some environments use sh
SHELL:=/bin/bash

# Determine the arch/os
ifeq (${XC_OS}, )
  XC_OS:=$(shell go env GOOS)
endif
export XC_OS

ifeq (${XC_ARCH}, )
  XC_ARCH:=$(shell go env GOARCH)
endif
export XC_ARCH

ARCH:=${XC_OS}_${XC_ARCH}
export ARCH

ifeq (${IMAGE_ORG}, )
  IMAGE_ORG=openebs
  export IMAGE_ORG
endif

# Specify the date of build
DBUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Specify the docker arg for repository url
ifeq (${DBUILD_REPO_URL}, )
  DBUILD_REPO_URL="https://github.com/openebs/external-storage"
  export DBUILD_REPO_URL
endif

# Specify the docker arg for website url
ifeq (${DBUILD_SITE_URL}, )
  DBUILD_SITE_URL="https://openebs.io"
  export DBUILD_SITE_URL
endif

export DBUILD_ARGS=--build-arg DBUILD_DATE=${DBUILD_DATE} --build-arg DBUILD_REPO_URL=${DBUILD_REPO_URL} --build-arg DBUILD_SITE_URL=${DBUILD_SITE_URL} --build-arg ARCH=${ARCH}


ifeq ($(REGISTRY),)
	REGISTRY = ${IMAGE_ORG}/
endif
ifeq ($(VERSION),)
	VERSION = latest
endif

ifeq ($(ARCH),linux_arm64)
  IMAGE_CONTROLLER = $(REGISTRY)snapshot-controller-arm64:$(VERSION)
  MUTABLE_IMAGE_CONTROLLER = $(REGISTRY)snapshot-controller-arm64:latest
  IMAGE_PROVISIONER = $(REGISTRY)snapshot-provisioner-arm64:$(VERSION)
  MUTABLE_IMAGE_PROVISIONER = $(REGISTRY)snapshot-provisioner-arm64:latest
else
  IMAGE_CONTROLLER = $(REGISTRY)snapshot-controller:$(VERSION)
  MUTABLE_IMAGE_CONTROLLER = $(REGISTRY)snapshot-controller:latest
  IMAGE_PROVISIONER = $(REGISTRY)snapshot-provisioner:$(VERSION)
  MUTABLE_IMAGE_PROVISIONER = $(REGISTRY)snapshot-provisioner:latest
endif

.PHONY: all controller provisioner clean container container-quick push test

all: controller provisioner

controller:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o _output/bin/snapshot-controller cmd/snapshot-controller/snapshot-controller.go

provisioner:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o _output/bin/snapshot-provisioner cmd/snapshot-pv-provisioner/snapshot-pv-provisioner.go

clean:
	-rm -f deploy/docker/controller/snapshot-controller
	-rm -f deploy/docker/provisioner/snapshot-provisioner
	-rm -rf deploy/docker/controller/etc
	-rm -rf deploy/docker/controller/usr
	-rm -rf deploy/docker/provisioner/etc
	-rm -rf deploy/docker/provisioner/usr
	-rm -rf _output

test:
	go test `go list ./... | grep -v 'vendor'`

container: controller provisioner container-quick

container-quick:
	cp _output/bin/snapshot-controller deploy/docker/controller
	cp _output/bin/snapshot-provisioner deploy/docker/provisioner
	# Copy the root CA certificates -- cloudproviders need them
	cp -Rf deploy/ca-certificates/* deploy/docker/controller/.
	cp -Rf deploy/ca-certificates/* deploy/docker/provisioner/.
	docker build -t $(MUTABLE_IMAGE_CONTROLLER) ${DBUILD_ARGS} deploy/docker/controller
	docker tag $(MUTABLE_IMAGE_CONTROLLER) $(IMAGE_CONTROLLER)
	docker build -t $(MUTABLE_IMAGE_PROVISIONER) ${DBUILD_ARGS} deploy/docker/provisioner
	docker tag $(MUTABLE_IMAGE_PROVISIONER) $(IMAGE_PROVISIONER)

push: container
	docker push $(IMAGE_CONTROLLER)
	docker push $(MUTABLE_IMAGE_CONTROLLER)
	docker push $(IMAGE_PROVISIONER)
	docker push $(MUTABLE_IMAGE_PROVISIONER)
