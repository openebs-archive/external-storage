# Copyright 2020 The OpenEBS Authors. All rights reserved.
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

# Build openebs snapshot-provisioner docker images with buildx
# Experimental docker feature to build cross platform multi-architecture docker images
# https://docs.docker.com/buildx/working-with-buildx/

# build args
export DBUILD_ARGS=--build-arg DBUILD_DATE=${DBUILD_DATE} --build-arg DBUILD_REPO_URL=${DBUILD_REPO_URL} --build-arg DBUILD_SITE_URL=${DBUILD_SITE_URL} --build-arg RELEASE_TAG=${RELEASE_TAG}

ifeq (${TAG}, )
  export TAG=ci
endif

# default list of platforms for which multiarch image is built
ifeq (${PLATFORMS}, )
	export PLATFORMS="linux/amd64,linux/arm64"
endif

# if IMG_RESULT is unspecified, by default the image will be pushed to registry
ifeq (${IMG_RESULT}, load)
	export PUSH_ARG="--load"
    # if load is specified, image will be built only for the build machine architecture.
    export PLATFORMS="local"
else ifeq (${IMG_RESULT}, cache)
	# if cache is specified, image will only be available in the build cache, it won't be pushed or loaded
	# therefore no PUSH_ARG will be specified
else
	export PUSH_ARG="--push"
endif

# Name of the multiarch image for snapshot-provisioner and controller
DOCKERX_IMAGE_SNAPSHOTTER:=${IMAGE_ORG}/snapshot-controller:${TAG}
DOCKERX_IMAGE_PROVISIONER:=${IMAGE_ORG}/snapshot-provisioner:${TAG}

.PHONY: docker.buildx
docker.buildx:
	export DOCKER_CLI_EXPERIMENTAL=enabled
	@if ! docker buildx ls | grep -q container-builder; then\
		docker buildx create --platform ${PLATFORMS} --name container-builder --use;\
	fi
	@docker buildx build --platform "${PLATFORMS}" \
		-t "$(DOCKERX_IMAGE_NAME)" ${BUILD_ARGS} \
		-f $(PWD)/snapshot/deploy/docker/$(COMPONENT)/$(COMPONENT).Dockerfile \
		. ${PUSH_ARG}
	@echo "--> Build docker image: $(DOCKERX_IMAGE_NAME)"
	@echo

.PHONY: docker.buildx.snapshot-controller
docker.buildx.snapshot-controller: DOCKERX_IMAGE_NAME=$(DOCKERX_IMAGE_SNAPSHOTTER)
docker.buildx.snapshot-controller: COMPONENT=controller
docker.buildx.snapshot-controller: BUILD_ARGS=$(DBUILD_ARGS)
docker.buildx.snapshot-controller: docker.buildx

.PHONY: docker.buildx.snapshot-provisioner
docker.buildx.snapshot-provisioner: DOCKERX_IMAGE_NAME=$(DOCKERX_IMAGE_PROVISIONER)
docker.buildx.snapshot-provisioner: COMPONENT=provisioner
docker.buildx.snapshot-provisioner: BUILD_ARGS=$(DBUILD_ARGS)
docker.buildx.snapshot-provisioner: docker.buildx


.PHONY: buildx.push.snapshot-controller
buildx.push.snapshot-controller:
	BUILDX=true DIMAGE=${IMAGE_ORG}/snapshot-controller ./buildxpush

.PHONY: buildx.push.snapshot-provisioner
buildx.push.snapshot-provisioner:
	BUILDX=true DIMAGE=${IMAGE_ORG}/snapshot-provisioner ./buildxpush
