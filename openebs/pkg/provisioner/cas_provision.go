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
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/kubernetes-incubator/external-storage/lib/util"
	"github.com/kubernetes-incubator/external-storage/openebs/pkg/apis/openebs.io/v1alpha1"
	mvol "github.com/kubernetes-incubator/external-storage/openebs/pkg/volume"
	mv1alpha1 "github.com/kubernetes-incubator/external-storage/openebs/pkg/volume/v1alpha1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type openEBSCASProvisioner struct {
	// Maya-API Server URI running in the cluster
	endpoint string

	// Identity of this openEBSProvisioner, set to node's name. Used to identify
	// "this" provisioner's PVs.
	identity string
}

// NewOpenEBSProvisioner creates a new openebs provisioner
func NewOpenEBSCASProvisioner(client kubernetes.Interface) (controller.Provisioner, error) {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		return nil, fmt.Errorf("Env variable 'NODE_NAME' is not set")
	}
	var openebsObj mvol.OpenEBSVolume
	//Get maya-apiserver IP address from cluster
	addr, err := openebsObj.GetMayaClusterIP(client)

	if err != nil {
		glog.Errorf("Error getting maya-apiserver IP Address: %v", err)
		return nil, err
	}
	mayaServiceURI := "http://" + addr + ":5656"

	//Set maya-apiserver IP address along with default port
	os.Setenv("MAPI_ADDR", mayaServiceURI)

	return &openEBSCASProvisioner{
		identity: nodeName,
		endpoint: mayaServiceURI,
	}, nil
}

var _ controller.Provisioner = &openEBSCASProvisioner{}

// Provision creates a storage asset and returns a PV object representing it.
func (p *openEBSCASProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {

	//Issue a request to Maya API Server to create a volume
	var openebsCASVol mv1alpha1.CASVolume
	casVolume := v1alpha1.CASVolume{}

	volSize := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	casVolume.Spec.Capacity = volSize.String()

	className := GetStorageClassName(options)

	// creating a map b/c have to initialize the map using the make function before
	// adding any elements to avoid nil map assignment error
	mapLabels := make(map[string]string)

	if className == nil {
		glog.Errorf("Volume has no storage class specified")
	} else {
		mapLabels[string(v1alpha1.StorageClassKey)] = *className
		casVolume.Labels = mapLabels
	}

	casVolume.Labels[string(v1alpha1.NamespaceKey)] = options.PVC.Namespace
	casVolume.Namespace = options.PVC.Namespace
	casVolume.Labels[string(v1alpha1.PersistentVolumeClaimKey)] = options.PVC.ObjectMeta.Name
	casVolume.Name = options.PVName

	// Check if volume already exists
	// if present then return the read values
	// if unexpected error then return the error
	// if absent then create volume
	glog.V(2).Infof("Checking if volume %q already exists", options.PVName)
	err := openebsCASVol.ReadVolume(options.PVName, options.PVC.Namespace, *className, &casVolume)
	if err == nil {
		glog.V(2).Infof("Volume %q already present", options.PVName)
	} else if err.Error() != http.StatusText(404) {
		// any error other than 404 is unexpected error
		glog.Errorf("Unexpected error occurred while trying to read the volume: %s", err)
		return nil, err
	} else if err.Error() == http.StatusText(404) {
		// Create the volume and read it
		glog.V(2).Infof("Volume %q does not exist,attempting to create volume", options.PVName)
		err = openebsCASVol.CreateVolume(casVolume)
		if err != nil {
			glog.Errorf("Failed to create volume:  %+v, error: %s", options, err.Error())
			return nil, err
		}
		err = openebsCASVol.ReadVolume(options.PVName, options.PVC.Namespace, *className, &casVolume)
		if err != nil {
			glog.Errorf("Failed to read volume: %v", err)
			return nil, err
		}
		glog.V(2).Infof("VolumeInfo: created volume metadata : %#v", casVolume)
	}

	if !util.AccessModesContainedInAll(p.GetAccessModes(), options.PVC.Spec.AccessModes) {
		glog.Errorf("Invalid Access Modes: %v, Supported Access Modes: %v", options.PVC.Spec.AccessModes, p.GetAccessModes())
		return nil, fmt.Errorf("Invalid Access Modes: %v, Supported Access Modes: %v", options.PVC.Spec.AccessModes, p.GetAccessModes())
	}

	// Use annotations to specify the context using which the PV was created.
	volAnnotations := make(map[string]string)
	volAnnotations = Setlink(volAnnotations, options.PVName)
	volAnnotations["openEBSProvisionerIdentity"] = p.identity
	volAnnotations["openebs.io/cas-type"] = casVolume.Spec.CasType

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.PVName,
			Annotations: volAnnotations,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			MountOptions:                  options.MountOptions,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				ISCSI: &v1.ISCSIPersistentVolumeSource{
					TargetPortal: casVolume.Spec.TargetPortal,
					IQN:          casVolume.Spec.Iqn,
					Lun:          casVolume.Spec.Lun,
					FSType:       casVolume.Spec.FSType,
					ReadOnly:     false,
				},
			},
		},
	}
	return pv, nil
}

func (p *openEBSCASProvisioner) GetAccessModes() []v1.PersistentVolumeAccessMode {
	return []v1.PersistentVolumeAccessMode{
		v1.ReadWriteOnce,
	}
}

// Delete removes the storage asset that was created by Provision represented
// by the given PV.
func (p *openEBSCASProvisioner) Delete(volume *v1.PersistentVolume) error {

	var openebsCASVol mv1alpha1.CASVolume
	_, ok := volume.Annotations["openEBSProvisionerIdentity"]
	if !ok {
		return errors.New("identity annotation not found on PV")
	}

	// openebs provisioner is stateless container, can be reschuduled in any
	// other node,so having identity check is not relvent here unless we use
	// some other identity check.
	// TODO use provisionerName instead of Node name
	/*
		if ann != p.identity {
			return &controller.IgnoredError{Reason: "identity annotation on PV does not match ours"}
		}
	*/

	// Issue a delete request to Maya API Server
	err := openebsCASVol.DeleteVolume(volume.Name, volume.Spec.ClaimRef.Namespace)
	if err != nil {
		glog.Errorf("Failed to delete volume %s, error: %s", volume, err.Error())
		return err
	}

	return nil
}
