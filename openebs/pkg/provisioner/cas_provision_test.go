/*
Copyright 2018 The Kubernetes and OpenEBS Authors.

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
	"reflect"
	"testing"

	"github.com/kubernetes-incubator/external-storage/openebs/pkg/apis/openebs.io/v1alpha1"
)

func Test_getVolAnnotations(t *testing.T) {
	tests := map[string]struct {
		annotations map[string]string
		casType     string
		want        map[string]string
	}{
		"All Annotations are present in CASVolume object and castype is set": {
			annotations: map[string]string{
				string(v1alpha1.CASTemplateKeyForVolumeCreate):   "createVolCAST",
				string(v1alpha1.CASTemplateKeyForVolumeDelete):   "deleteVolCAST",
				string(v1alpha1.CASTemplateKeyForVolumeRead):     "readVolCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotCreate): "createSnapCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotRead):   "readSnapCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotDelete): "deleteSnapCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotList):   "listSnapCAST",
			},
			casType: "jiva",
			want: map[string]string{
				string(v1alpha1.CASTemplateKeyForVolumeCreate):   "createVolCAST",
				string(v1alpha1.CASTemplateKeyForVolumeDelete):   "deleteVolCAST",
				string(v1alpha1.CASTemplateKeyForVolumeRead):     "readVolCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotCreate): "createSnapCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotRead):   "readSnapCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotDelete): "deleteSnapCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotList):   "listSnapCAST",
				string(v1alpha1.CASTypeKey):                      "jiva",
			},
		},
		"Some Annotations are absent in CASVolume object and castype is set": {
			annotations: map[string]string{
				string(v1alpha1.CASTemplateKeyForVolumeCreate): "createCAST",
				string(v1alpha1.CASTemplateKeyForVolumeRead):   "readCAST",
			},
			casType: "jiva",
			want: map[string]string{
				string(v1alpha1.CASTemplateKeyForVolumeCreate):   "createCAST",
				string(v1alpha1.CASTemplateKeyForVolumeRead):     "readCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotCreate): "",
				string(v1alpha1.CASTemplateKeyForSnapshotRead):   "",
				string(v1alpha1.CASTemplateKeyForSnapshotDelete): "",
				string(v1alpha1.CASTemplateKeyForSnapshotList):   "",
				string(v1alpha1.CASTemplateKeyForVolumeDelete):   "",
				string(v1alpha1.CASTypeKey):                      "jiva",
			},
		},
		"Some Extra Annotations are present in CASVolume object and castype is set": {
			annotations: map[string]string{
				string(v1alpha1.CASTemplateKeyForVolumeCreate):   "createCAST",
				string(v1alpha1.CASTemplateKeyForVolumeRead):     "readCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotCreate): "createSnapCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotRead):   "readSnapCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotDelete): "deleteSnapCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotList):   "listSnapCAST",
				"extraAnnotation1":                               "val1",
				"extraAnnotation2":                               "val2",
			},
			casType: "cstor",
			want: map[string]string{
				string(v1alpha1.CASTemplateKeyForVolumeCreate):   "createCAST",
				string(v1alpha1.CASTemplateKeyForVolumeRead):     "readCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotCreate): "createSnapCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotRead):   "readSnapCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotDelete): "deleteSnapCAST",
				string(v1alpha1.CASTemplateKeyForSnapshotList):   "listSnapCAST",
				string(v1alpha1.CASTemplateKeyForVolumeDelete):   "",
				string(v1alpha1.CASTypeKey):                      "cstor",
			},
		},
		"Annotations are missing in CASVolume object and castype is empty": {
			annotations: map[string]string{},
			casType:     "",
			want: map[string]string{
				string(v1alpha1.CASTemplateKeyForVolumeCreate):   "",
				string(v1alpha1.CASTemplateKeyForVolumeRead):     "",
				string(v1alpha1.CASTemplateKeyForVolumeDelete):   "",
				string(v1alpha1.CASTemplateKeyForSnapshotCreate): "",
				string(v1alpha1.CASTemplateKeyForSnapshotRead):   "",
				string(v1alpha1.CASTemplateKeyForSnapshotDelete): "",
				string(v1alpha1.CASTemplateKeyForSnapshotList):   "",
				string(v1alpha1.CASTypeKey):                      "",
			},
		},
		"Annotations field is nil in CASVolume object and castype is set": {
			annotations: nil,
			casType:     "cstor",
			want: map[string]string{
				string(v1alpha1.CASTypeKey): "cstor",
			},
		},
	}
	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			casVolume := v1alpha1.CASVolume{}
			casVolume.Spec.CasType = mock.casType
			casVolume.Annotations = mock.annotations
			if got := getVolAnnotations(casVolume); !reflect.DeepEqual(got, mock.want) {
				t.Errorf("getVolAnnotations() = %v, want %v", got, mock.want)
			}
		})
	}
}
