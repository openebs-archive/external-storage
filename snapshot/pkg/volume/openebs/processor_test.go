package openebs

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetNameAndNameSpaceFromSnapshoName(t *testing.T) {
	cases := map[string]struct {
		name               string
		expectErr          error
		expectNamespace    string
		expectSnapshotName string
	}{

		"SnapshotName with Namespace":    {"percona/fastfurious", nil, "percona", "fastfurious"},
		"SnapshotName without Namespace": {"fastfurious", fmt.Errorf("invalid snapshot name"), "", ""},
		"Invalid Unique SnapshotName":    {"k8s/percona/fastfurious", fmt.Errorf("invalid snapshot name"), "", ""},
		"Nil SnapshotName":               {"", fmt.Errorf("invalid snapshot name"), "", ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			namespace, snapshotName, err := GetNameAndNameSpaceFromSnapshotName(tc.name)

			if !reflect.DeepEqual(err, tc.expectErr) {
				t.Errorf("Expected %v, got %v", tc.expectErr, err)
			}
			if !reflect.DeepEqual(namespace, tc.expectNamespace) {
				t.Errorf("Expected %v, got %v", tc.expectNamespace, namespace)
			}
			if !reflect.DeepEqual(snapshotName, tc.expectSnapshotName) {
				t.Errorf("Expected %v, got %v", tc.expectSnapshotName, snapshotName)
			}
		})
	}
}
