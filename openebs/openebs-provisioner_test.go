package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseClassParameters(t *testing.T) {
	//	cfgs := make(map[string]string)

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
