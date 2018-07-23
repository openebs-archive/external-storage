/*
Copyright 2017 The Kubernetes Authors.

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
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/kubernetes-incubator/external-storage/lib/util"
	mvol "github.com/kubernetes-incubator/external-storage/openebs/pkg/volume"
	mayav1 "github.com/kubernetes-incubator/external-storage/openebs/types/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	provisionerName = "openebs.io/provisioner-iscsi"
)

type openEBSProvisioner struct {
	// Maya-API Server URI running in the cluster
	mapiURI string

	// Identity of this openEBSProvisioner, set to node's name. Used to identify
	// "this" provisioner's PVs.
	identity string
}

// NewOpenEBSProvisioner creates a new openebs provisioner
func NewOpenEBSProvisioner(client kubernetes.Interface) controller.Provisioner {

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

	return &openEBSProvisioner{
		mapiURI:  mayaServiceURI,
		identity: nodeName,
	}
}

var _ controller.Provisioner = &openEBSProvisioner{}

// Provision creates a storage asset and returns a PV object representing it.
func (p *openEBSProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {

	//Issue a request to Maya API Server to create a volume
	var volume mayav1.Volume
	var openebsVol mvol.OpenEBSVolume
	volumeSpec := mayav1.VolumeSpec{}

	volSize := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	volumeSpec.Metadata.Labels.Storage = volSize.String()

	className := GetStorageClassName(options)

	if className == nil {
		glog.Errorf("Volume has no storage class specified")
	} else {
		volumeSpec.Metadata.Labels.StorageClass = *className
	}
	volumeSpec.Metadata.Labels.Namespace = options.PVC.Namespace
	volumeSpec.Metadata.Labels.PersistentVolumeClaim = options.PVC.ObjectMeta.Name
	volumeSpec.Metadata.Name = options.PVName

	//Pass through labels from PVC to maya-apiserver
	volumeSpec.Metadata.Labels.Application = options.PVC.ObjectMeta.GetLabels()[mayav1.PVCLabelsApplication]
	volumeSpec.Metadata.Labels.ReplicaTopoKeyDomain = options.PVC.ObjectMeta.GetLabels()[mayav1.PVCLabelsReplicaTopKeyDomain]
	volumeSpec.Metadata.Labels.ReplicaTopoKeyType = options.PVC.ObjectMeta.GetLabels()[mayav1.PVCLabelsReplicaTopKeyType]

	err := openebsVol.CreateVolume(volumeSpec)
	if err != nil {
		glog.Errorf("Error creating volume: %v", err)
		return nil, err
	}

	err = openebsVol.ListVolume(options.PVName, options.PVC.Namespace, &volume)
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

	fsType, err := ParseClassParameters(options.Parameters)
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
			MountOptions:                  options.MountOptions,
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

func (p *openEBSProvisioner) GetAccessModes() []v1.PersistentVolumeAccessMode {
	return []v1.PersistentVolumeAccessMode{
		v1.ReadWriteOnce,
	}
}

// The following will be used by the dashboard, to display links on PV page
func Setlink(volAnnotations map[string]string, pvName string) map[string]string {
	userLinks := make([]string, 0)
	localMonitoringURL := os.Getenv("OPENEBS_MONITOR_URL")
	if localMonitoringURL != "" {
		localMonitorLinkName := os.Getenv("OPENEBS_MONITOR_LINK_NAME")
		if localMonitorLinkName == "" {
			localMonitorLinkName = "monitor"
		}
		localMonitorVolKey := os.Getenv("OPENEBS_MONITOR_VOLKEY")
		if localMonitorVolKey != "" {
			localMonitoringURL += localMonitorVolKey + "=" + pvName
		}
		userLinks = append(userLinks, "\""+localMonitorLinkName+"\":\""+localMonitoringURL+"\"")
	}
	mayaPortalURL := os.Getenv("MAYA_PORTAL_URL")
	if mayaPortalURL != "" {
		mayaPortalLinkName := os.Getenv("MAYA_PORTAL_LINK_NAME")
		if mayaPortalLinkName == "" {
			mayaPortalLinkName = "maya"
		}
		userLinks = append(userLinks, "\""+mayaPortalLinkName+"\":\""+mayaPortalURL+"\"")
	}
	if len(userLinks) > 0 {
		volAnnotations["alpha.dashboard.kubernetes.io/links"] = "{" + strings.Join(userLinks, ",") + "}"
	}

	return volAnnotations
}

// Delete removes the storage asset that was created by Provision represented
// by the given PV.
func (p *openEBSProvisioner) Delete(volume *v1.PersistentVolume) error {

	var openebsVol mvol.OpenEBSVolume

	ann, ok := volume.Annotations["openEBSProvisionerIdentity"]
	if !ok {
		return errors.New("identity annotation not found on PV")
	}
	if ann != p.identity {
		return &controller.IgnoredError{Reason: "identity annotation on PV does not match ours"}
	}

	// Issue a delete request to Maya API Server
	err := openebsVol.DeleteVolume(volume.Name, volume.Spec.ClaimRef.Namespace)
	if err != nil {
		glog.Errorf("Failed to delete volume %s, error: %s", volume, err.Error())
		return err
	}

	return nil
}
