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

ifeq (${ARCH},linux_arm64)
  DIMAGE:="openebs/openebs-k8s-provisioner-arm64"
else
  DIMAGE:="openebs/openebs-k8s-provisioner"
endif
export DIMAGE

.PHONY: image clean build push deploy
.DEFAULT_GOAL := build


build:
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o openebs-provisioner ./cmd/openebs-provisioner/

image: build
	@cp openebs-provisioner buildscripts/docker/
	@cd buildscripts/docker && sudo docker build -t ${DIMAGE}:ci --build-arg BASE_IMAGE=${BASEIMAGE} .

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

