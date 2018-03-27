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

package openebs

import (
	"fmt"
	"time"

	mApiv1 "github.com/kubernetes-incubator/external-storage/openebs/pkg/v1"
	mayav1 "github.com/kubernetes-incubator/external-storage/openebs/types/v1"

	"github.com/golang/glog"
	crdv1 "github.com/kubernetes-incubator/external-storage/snapshot/pkg/apis/crd/v1"
	"github.com/kubernetes-incubator/external-storage/snapshot/pkg/cloudprovider"
	"github.com/kubernetes-incubator/external-storage/snapshot/pkg/volume"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	openEBSPersistentDiskPluginName = "openebs"
)

type openEBSPlugin struct {
	OpenEBSVolume
}

var _ volume.Plugin = &openEBSPlugin{}

// RegisterPlugin registers the volume plugin
func RegisterPlugin() volume.Plugin {
	return &openEBSPlugin{}
}

// GetPluginName gets the name of the volume plugin
func GetPluginName() string {
	return "openebs"
}

func (h *openEBSPlugin) Init(_ cloudprovider.Interface) {
}

func (h *openEBSPlugin) SnapshotCreate(pv *v1.PersistentVolume, tags *map[string]string) (*crdv1.VolumeSnapshotDataSource, *[]crdv1.VolumeSnapshotCondition, error) {
	spec := &pv.Spec
	if spec == nil || spec.ISCSI == nil {
		return nil, nil, fmt.Errorf("invalid PV spec %v", spec)
	}

	//snapshotName := &tags["kubernetes.io/created-for/snapshot/name"]
	snapshotName := createSnapshotName(pv.Name)

	fmt.Println("PV Name , snapshot Name", pv.Name, snapshotName)
	_, err := h.CreateSnapshot(pv.Name, snapshotName)
	if err != nil {
		glog.Errorf("failed to create snapshot for volume :%v, err: %v", pv.Name, err)
	}
	glog.V(1).Info("snapshot %v created successfully", snapshotName)

	cond := []crdv1.VolumeSnapshotCondition{}
	if err == nil {
		cond = []crdv1.VolumeSnapshotCondition{
			{
				Status:             v1.ConditionTrue,
				Message:            "Snapshot created successfully",
				LastTransitionTime: metav1.Now(),
				Type:               crdv1.VolumeSnapshotConditionReady,
			},
		}
	} else {
		glog.V(2).Infof("failed to create snapshot, err: %v", err)
		cond = []crdv1.VolumeSnapshotCondition{
			{
				Status:             v1.ConditionTrue,
				Message:            fmt.Sprintf("Failed to create the snapshot: %v", err),
				LastTransitionTime: metav1.Now(),
				Type:               crdv1.VolumeSnapshotConditionError,
			},
		}
	}

	res := &crdv1.VolumeSnapshotDataSource{
		OpenEBSSnapshot: &crdv1.OpenEBSVolumeSnapshotSource{
			SnapshotID: snapshotName,
		},
	}
	return res, &cond, err
}

func createSnapshotName(pvName string) string {
	name := pvName + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
	return name
}

func (h *openEBSPlugin) SnapshotDelete(src *crdv1.VolumeSnapshotDataSource, pv *v1.PersistentVolume) error {
	if src == nil || src.OpenEBSSnapshot == nil {
		return fmt.Errorf("invalid VolumeSnapshotDataSource: %v", src)
	}
	snapshotID := src.OpenEBSSnapshot.SnapshotID
	glog.V(1).Infof("Received snapshot :%v delete request", snapshotID)

	_, err := h.DeleteSnapshot(snapshotID)
	if err != nil {
		glog.Errorf("failed to delete snapshot: %v, err: %v", snapshotID, err)
	}

	glog.V(1).Infof("snapshot deleted :%v successfully", snapshotID)
	return err
}

func (h *openEBSPlugin) DescribeSnapshot(snapshotData *crdv1.VolumeSnapshotData) (snapConditions *[]crdv1.VolumeSnapshotCondition, isCompleted bool, err error) {
	if snapshotData == nil || snapshotData.Spec.OpenEBSSnapshot == nil {
		return nil, false, fmt.Errorf("failed to retrieve Snapshot spec")
	}

	snapshotID := snapshotData.Spec.OpenEBSSnapshot.SnapshotID
	glog.V(1).Infof("received describe request on snapshot:%v", snapshotID)

	// TODO implement snapshot-info based on snapshotID
	resp, err := h.SnapshotInfo(snapshotData.Spec.PersistentVolumeRef.Name, snapshotID)

	if err != nil {
		glog.Errorf("failed to describe snapshot:%v", snapshotID)
	}

	glog.V(1).Infof("snapshot details:%v", string(resp))

	if len(snapshotData.Status.Conditions) == 0 {
		return nil, false, fmt.Errorf("No status condtions in VoluemSnapshotData for openebs snapshot type")
	}

	lastCondIdx := len(snapshotData.Status.Conditions) - 1
	retCondType := crdv1.VolumeSnapshotConditionError

	switch snapshotData.Status.Conditions[lastCondIdx].Type {
	case crdv1.VolumeSnapshotDataConditionReady:
		retCondType = crdv1.VolumeSnapshotConditionReady
	case crdv1.VolumeSnapshotDataConditionPending:
		retCondType = crdv1.VolumeSnapshotConditionPending
		// Error out.
	}
	retCond := []crdv1.VolumeSnapshotCondition{
		{
			Status:             snapshotData.Status.Conditions[lastCondIdx].Status,
			Message:            snapshotData.Status.Conditions[lastCondIdx].Message,
			LastTransitionTime: snapshotData.Status.Conditions[lastCondIdx].LastTransitionTime,
			Type:               retCondType,
		},
	}
	return &retCond, true, nil
}

// FindSnapshot finds a VolumeSnapshot by matching metadata
func (h *openEBSPlugin) FindSnapshot(tags *map[string]string) (*crdv1.VolumeSnapshotDataSource, *[]crdv1.VolumeSnapshotCondition, error) {
	glog.Infof("FindSnapshot by tags: %#v", *tags)

	// TODO: Implement FindSnapshot
	return &crdv1.VolumeSnapshotDataSource{
		OpenEBSSnapshot: &crdv1.OpenEBSVolumeSnapshotSource{
			SnapshotID: "",
		},
	}, nil, nil
}

// SnapshotRestore restore to any created snapshot
func (h *openEBSPlugin) SnapshotRestore(snapshotData *crdv1.VolumeSnapshotData,
	pvc *v1.PersistentVolumeClaim,
	pvName string,
	parameters map[string]string,
) (*v1.PersistentVolumeSource, map[string]string, error) {

	if snapshotData == nil || snapshotData.Spec.OpenEBSSnapshot == nil {
		return nil, nil, fmt.Errorf("Invalid Snapshot spec")
	}
	if pvc == nil {
		return nil, nil, fmt.Errorf("Invalid PVC spec")
	}

	// restore snapshot to a PV
	snapshotID := snapshotData.Spec.OpenEBSSnapshot.SnapshotID
	pvRefName := snapshotData.Spec.PersistentVolumeRef.Name

	var oldvolume, newvolume mayav1.Volume
	var openebsVol mApiv1.OpenEBSVolume
	volumeSpec := mayav1.VolumeSpec{}

	volSize := pvc.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	volumeSpec.Metadata.Labels.Storage = volSize.String()
	volumeSpec.Metadata.Labels.StorageClass = *pvc.Spec.StorageClassName
	volumeSpec.Metadata.Labels.Namespace = pvc.Namespace
	volumeSpec.Metadata.Name = pvName

	err := openebsVol.ListVolume(pvRefName, &oldvolume)
	if err != nil {
		glog.Errorf("Error getting volume details: %v", err)
		return nil, nil, err
	}
	var cloneIP string
	for key, value := range oldvolume.Metadata.Annotations.(map[string]interface{}) {
		switch key {
		case "vsm.openebs.io/controller-ips":
			cloneIP = value.(string)
		}
	}
	volumeSpec.CloneIP = cloneIP
	volumeSpec.SnapshotName = snapshotID

	_, err = openebsVol.CreateVolume(volumeSpec)
	if err != nil {
		glog.Errorf("Error creating volume: %v", err)
		return nil, nil, err
	}
	err = openebsVol.ListVolume(pvName, &newvolume)
	if err != nil {
		glog.Errorf("Error getting volume details: %v", err)
		return nil, nil, err
	}

	var iqn, targetPortal string

	for key, value := range newvolume.Metadata.Annotations.(map[string]interface{}) {
		switch key {
		case "vsm.openebs.io/iqn":
			iqn = value.(string)
		case "vsm.openebs.io/targetportals":
			targetPortal = value.(string)
		}
	}

	if err != nil {
		glog.Errorf("snapshot :%v restore failed, err:%v", snapshotID, err)
		return nil, nil, fmt.Errorf("failed to restore %s, err: %v", snapshotID, err)
	}

	glog.V(1).Infof("snapshot restored successfully to: %v", snapshotID)

	pv := &v1.PersistentVolumeSource{
		ISCSI: &v1.ISCSIVolumeSource{
			TargetPortal: targetPortal,
			IQN:          iqn,
			Lun:          1,
			FSType:       "ext4",
			ReadOnly:     false,
		},
	}
	return pv, nil, nil
}

func (h *openEBSPlugin) VolumeDelete(pv *v1.PersistentVolume) error {
	if pv == nil || pv.Spec.ISCSI == nil {
		return fmt.Errorf("invalid VolumeSnapshotDataSource: %v", pv)
	}
	var openebsVol mApiv1.OpenEBSVolume

	// Issue a delete request to Maya API Server
	err := openebsVol.DeleteVolume(pv.Name)
	if err != nil {
		glog.Errorf("Error while deleting volume: %v", err)
		return err
	}
	return nil
}
