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

#$DST_REPO/external-storage/openebs/ci/helm_install_openebs.sh
#rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi

# Calling snapshot test
$DST_REPO/external-storage/openebs/ci/snapshot/snapshot_deploy.sh
rc=$?; if [[ $rc != 0 ]]; then exit $rc; fi


