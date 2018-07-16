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

package volume

import (
	"errors"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	mApiv1 "github.com/kubernetes-incubator/external-storage/openebs/pkg/v1"
	"k8s.io/api/core/v1"
)

// Delete removes the storage asset that was created by Provision represented
// by the given PV.
func (p *openEBSProvisionerV1alpha1) Delete(volume *v1.PersistentVolume) error {

	var openebsCASVol mApiv1.OpenEBSVolumeV1Alpha1
	ann, ok := volume.Annotations["openEBSProvisionerIdentity"]
	if !ok {
		return errors.New("identity annotation not found on PV")
	}
	if ann != p.identity {
		return &controller.IgnoredError{Reason: "identity annotation on PV does not match ours"}
	}

	// Issue a delete request to Maya API Server
	err := openebsCASVol.DeleteVolume(volume.Name, volume.Spec.ClaimRef.Namespace)
	if err != nil {
		glog.Errorf("Failed to delete volume %s, error: %s", volume, err.Error())
		return err
	}

	return nil
}
