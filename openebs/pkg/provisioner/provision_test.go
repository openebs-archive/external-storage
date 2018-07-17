/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provisioner

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestParseClassParameters(t *testing.T) {
	cases := map[string]struct {
		cfgs         map[string]string
		expectErr    error
		expectfstype string
	}{
		"Dafault fstype": {
			cfgs:         map[string]string{"openebs.io/fstype": ""},
			expectErr:    nil,
			expectfstype: "ext4",
		},
		"ext4 fstype": {
			cfgs:         map[string]string{"openebs.io/fstype": "ext4"},
			expectErr:    nil,
			expectfstype: "ext4",
		},
		"xfs fstype": {
			cfgs:         map[string]string{"openebs.io/fstype": "xfs"},
			expectErr:    nil,
			expectfstype: "xfs",
		},
		"Invalid fstype": {
			cfgs:         map[string]string{"openebs.io/fstype": "nfs"},
			expectErr:    fmt.Errorf("Filesystem nfs is not supported"),
			expectfstype: "",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			fstype, err := parseClassParameters(tc.cfgs)
			if !reflect.DeepEqual(err, tc.expectErr) {
				t.Errorf("Expected %v, got %v", tc.expectErr, err)
			}
			if !reflect.DeepEqual(fstype, tc.expectfstype) {
				t.Errorf("Expected %v, got %v", tc.expectfstype, fstype)
			}
		})
	}
}

func TestParseClassENVParameters(t *testing.T) {
	cases := map[string]struct {
		cfgs         map[string]string
		expectErr    error
		expectfstype string
	}{
		"Dafault fstype": {
			cfgs:         map[string]string{"openebs.io/fstype": ""},
			expectErr:    nil,
			expectfstype: "ext4",
		},
		"ext4 fstype": {
			cfgs:         map[string]string{"openebs.io/fstype": "ext4"},
			expectErr:    nil,
			expectfstype: "ext4",
		},
		"ENV fstype nfs": {
			cfgs:         map[string]string{"openebs.io/fstype": "nfs"},
			expectErr:    nil,
			expectfstype: "nfs",
		},
		"Invalid fstype": {
			cfgs:         map[string]string{"openebs.io/fstype": "abcd"},
			expectErr:    fmt.Errorf("Filesystem abcd is not supported"),
			expectfstype: "",
		},
		"ENV fstype btrfs": {
			cfgs:         map[string]string{"openebs.io/fstype": "btrfs"},
			expectErr:    nil,
			expectfstype: "btrfs",
		},
		"ENV fstype zfs": {
			cfgs:         map[string]string{"openebs.io/fstype": "zfs"},
			expectErr:    nil,
			expectfstype: "zfs",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			os.Setenv("OPENEBS_VALID_FSTYPE", "nfs,btrfs,zfs")
			defer os.Unsetenv("OPENEBS_VALID_FSTYPE")

			fstype, err := parseClassParameters(tc.cfgs)
			if !reflect.DeepEqual(err, tc.expectErr) {
				t.Errorf("Expected %v, got %v", tc.expectErr, err)
			}
			if !reflect.DeepEqual(fstype, tc.expectfstype) {
				t.Errorf("Expected %v, got %v", tc.expectfstype, fstype)
			}
		})
	}
}
