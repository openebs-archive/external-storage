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

ifeq (${BASE_DOCKER_IMAGEARM64}, )
  BASE_DOCKER_IMAGEARM64 = "arm64v8/ubuntu:18.04"
  export BASE_DOCKER_IMAGEARM64
endif

.PHONY: image clean build push deploy
.DEFAULT_GOAL := build



build:
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o openebs-provisioner ./cmd/openebs-provisioner/

push:
	@cp openebs-provisioner buildscripts/docker/
	@cd buildscripts/docker && sudo docker build -t openebs/openebs-k8s-provisioner:ci .

image: build push

deploy:
	@cp openebs-provisioner buildscripts/docker/
	@cd buildscripts/docker && sudo docker build -t ${DIMAGE}:ci .
	@sh buildscripts/push

clean:
	rm -rf vendor
	rm -f openebs-provisioner
	rm -f buildscripts/docker/openebs-provisioner

.PHONY: image.arm64
image.arm64: build
	@cp openebs-provisioner buildscripts/docker/
	@cd buildscripts/docker && sudo docker build -f Dockerfile.arm64 -t openebs/openebs-k8s-provisioner-arm64:ci --build-arg BASE_IMAGE=${BASE_DOCKER_IMAGEARM64} .

