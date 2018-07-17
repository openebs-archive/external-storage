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
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/kubernetes-incubator/external-storage/lib/util"
	"github.com/kubernetes-incubator/external-storage/openebs/pkg/apis/v1alpha1"
	mutil "github.com/kubernetes-incubator/external-storage/openebs/pkg/util"
	mvol "github.com/kubernetes-incubator/external-storage/openebs/pkg/volume"
	mv1alpha1 "github.com/kubernetes-incubator/external-storage/openebs/pkg/volume/v1alpha1"
	mayav1 "github.com/kubernetes-incubator/external-storage/openebs/types/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type openEBSCASProvisioner struct {
	// Maya-API Server URI running in the cluster
	mapiURI string

	// Identity of this openEBSProvisioner, set to node's name. Used to identify
	// "this" provisioner's PVs.
	identity string
}

// NewOpenEBSProvisioner creates a new openebs provisioner
func NewOpenEBSCASProvisioner(client kubernetes.Interface) controller.Provisioner {

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		glog.Errorf("ENV variable 'NODE_NAME' is not set")
	}
	var openebsObj mvol.OpenEBSVolume
	//Get maya-apiserver IP address from cluster
	addr, err := openebsObj.GetMayaClusterIP(client)

	if err != nil {
		glog.Errorf("Error getting maya-apiserver IP Address: %v", err)
		return nil
	}
	mayaServiceURI := "http://" + addr + ":5656"

	//Set maya-apiserver IP address along with default port
	os.Setenv("MAPI_ADDR", mayaServiceURI)

	return &openEBSCASProvisioner{
		mapiURI:  mayaServiceURI,
		identity: nodeName,
	}
}

var _ controller.Provisioner = &openEBSCASProvisioner{}

// Provision creates a storage asset and returns a PV object representing it.
func (p *openEBSCASProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {

	//Issue a request to Maya API Server to create a volume
	var volume mayav1.Volume
	var openebsCASVol mv1alpha1.CASVolume
	casVolume := v1alpha1.CASVolume{}

	volSize := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	casVolume.Spec.Capacity = volSize.String()

	className := mutil.GetStorageClassName(options)

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

	_, err := openebsCASVol.CreateVolume(casVolume)
	if err != nil {
		glog.Errorf("Error creating volume: %v", err)
		return nil, err
	}

	err = openebsCASVol.ReadVolume(options.PVName, options.PVC.Namespace, &volume)
	if err != nil {
		glog.Errorf("Error getting volume details: %v", err)
		return nil, err
	}
	var iqn, targetPortal string

	for key, value := range volume.Metadata.Annotations.(map[string]interface{}) {
		switch key {
		case "vsm.openebs.io/iqn":
			iqn = value.(string)
		case "vsm.openebs.io/targetportals":
			targetPortal = value.(string)
		}
	}

	glog.V(2).Infof("Volume IQN: %v , Volume Target: %v", iqn, targetPortal)

	if !util.AccessModesContainedInAll(p.GetAccessModes(), options.PVC.Spec.AccessModes) {
		glog.Errorf("Invalid Access Modes: %v, Supported Access Modes: %v", options.PVC.Spec.AccessModes, p.GetAccessModes())
		return nil, fmt.Errorf("Invalid Access Modes: %v, Supported Access Modes: %v", options.PVC.Spec.AccessModes, p.GetAccessModes())
	}

	// Use annotations to specify the context using which the PV was created.
	volAnnotations := make(map[string]string)
	volAnnotations = Setlink(volAnnotations, options.PVName)
	volAnnotations["openEBSProvisionerIdentity"] = p.identity

	fsType, err := mutil.ParseClassParameters(options.Parameters)
	if err != nil {
		return nil, err
	}
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.PVName,
			Annotations: volAnnotations,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				ISCSI: &v1.ISCSIPersistentVolumeSource{
					TargetPortal: targetPortal,
					IQN:          iqn,
					Lun:          0,
					FSType:       fsType,
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
