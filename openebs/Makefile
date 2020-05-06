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

ifeq (${BASE_DOCKER_IMAGEARM64}, )
  BASE_DOCKER_IMAGEARM64 = "arm64v8/ubuntu:18.04"
  export BASE_DOCKER_IMAGEARM64
endif

ifeq (${ARCH},linux_arm64)
  BASEIMAGE:=${BASE_DOCKER_IMAGEARM64}
else
  # The ubuntu:16.04 image is being used as base image.
  BASEIMAGE:=ubuntu:16.04
endif
export BASEIMAGE

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


ifeq (${ARCH},linux_arm64)
  DIMAGE:="${IMAGE_ORG}/openebs-k8s-provisioner-arm64"
else
  DIMAGE:="${IMAGE_ORG}/openebs-k8s-provisioner"
endif
export DIMAGE

.PHONY: image clean build push deploy
.DEFAULT_GOAL := build


build:
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o openebs-provisioner ./cmd/openebs-provisioner/

image: build
	@cp openebs-provisioner buildscripts/docker/
	@cd buildscripts/docker && sudo docker build -t ${DIMAGE}:ci ${DBUILD_ARGS} --build-arg BASE_IMAGE=${BASEIMAGE} .

.PHONY: container
container: image

deploy:
	@cp openebs-provisioner buildscripts/docker/
	@cd buildscripts/docker && sudo docker build -t ${DIMAGE}:ci .
	@sh buildscripts/push

clean:
	rm -rf vendor
	rm -f openebs-provisioner
	rm -f buildscripts/docker/openebs-provisioner

