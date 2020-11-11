# Copyright 2020 The Kubernetes Authors.
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

FROM golang:1.14.7 as build

ARG RELEASE_TAG
ARG BRANCH
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""

ENV GO111MODULE=off \
  GOOS=${TARGETOS} \
  GOARCH=${TARGETARCH} \
  GOARM=${TARGETVARIANT} \
  DEBIAN_FRONTEND=noninteractive \
  PATH="/root/go/bin:${PATH}" \
  RELEASE_TAG=${RELEASE_TAG} \
  BRANCH=${BRANCH}

WORKDIR /go/src/github.com/kubernetes-incubator/external-storage

COPY . .

RUN make openebs

FROM ubuntu:16.04
RUN apt-get clean && rm -rf /var/lib/apt/lists/*
RUN apt-get update; exit 0

COPY --from=build /go/src/github.com/kubernetes-incubator/external-storage/openebs/openebs-provisioner /

ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL


LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.name="openebs-k8s-provisioner"
LABEL org.label-schema.description="OpenEBS Dynamic cStor and Jiva Provisioner"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

CMD ["/openebs-provisioner"]
