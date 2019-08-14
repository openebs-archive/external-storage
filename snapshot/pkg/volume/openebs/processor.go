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
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/openebs/pkg/apis/openebs.io/v1alpha1"
	"github.com/kubernetes-incubator/external-storage/openebs/pkg/provisioner"
	mvol_v1alpha1 "github.com/kubernetes-incubator/external-storage/openebs/pkg/volume/v1alpha1"
	crdv1 "github.com/kubernetes-incubator/external-storage/snapshot/pkg/apis/crd/v1"
	"github.com/kubernetes-incubator/external-storage/snapshot/pkg/cloudprovider"
	"github.com/kubernetes-incubator/external-storage/snapshot/pkg/volume"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	openEBSPersistentDiskPluginName = "openebs"
)

var (
	// SnapSupportedCASType is a map of supported volume type for snapshot
	// operations.
	SnapSupportedCASType = map[string]bool{
		"jiva":  true,
		"cstor": true,
	}
)

type openEBSPlugin struct {
	mvol_v1alpha1.CASVolume
}

var _ volume.Plugin = &openEBSPlugin{}

// RegisterPlugin registers the volume plugin
func RegisterPlugin() volume.Plugin {
	return &openEBSPlugin{}
}

func init() {
	// GetMayaService get the maya-service endpoint
	_ = GetMayaService()
}

// GetPluginName gets the name of the volume plugin
func GetPluginName() string {
	return "openebs"
}

func (h *openEBSPlugin) Init(_ cloudprovider.Interface) {
}

func (h *openEBSPlugin) SnapshotCreate(snapshot *crdv1.VolumeSnapshot, pv *v1.PersistentVolume, tags *map[string]string) (*crdv1.VolumeSnapshotDataSource, *[]crdv1.VolumeSnapshotCondition, error) {
	spec := &pv.Spec
	if spec == nil || spec.ISCSI == nil {
		return nil, nil, fmt.Errorf("invalid PV spec %v", spec)
	}

	snapObj := (*tags)["kubernetes.io/created-for/snapshot/name"]
	snapshotName := createSnapshotName(pv.Name, snapObj)

	casType := pv.Annotations["openebs.io/cas-type"]
	if casType == "" {
		casType = "jiva"
	}

	ok := SnapSupportedCASType[casType]
	if !ok {
		return nil, nil, fmt.Errorf("aborting create snapshot operation as specified volume type (%s) does not support snapshots", casType)
	}
	_, err := h.CreateSnapshot(casType, pv.Name, snapshotName, pv.Spec.ClaimRef.Namespace)
	if err != nil {
		glog.Errorf("failed to create snapshot for volume :%v, err: %v", pv.Name, err)
		return nil, nil, err
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

	sizeResource := pv.Spec.Capacity[v1.ResourceName(v1.ResourceStorage)]
	res := &crdv1.VolumeSnapshotDataSource{
		OpenEBSSnapshot: &crdv1.OpenEBSVolumeSnapshotSource{
			SnapshotID: snapshotName,
			Capacity:   sizeResource.String(),
		},
	}
	return res, &cond, err
}

func createSnapshotName(pvName string, snapObj string) string {
	name := pvName + "_" + snapObj + "_" + fmt.Sprintf("%d", time.Now().UnixNano())

	return name
}

func (h *openEBSPlugin) SnapshotDelete(src *crdv1.VolumeSnapshotDataSource, pv *v1.PersistentVolume) error {
	if src == nil || src.OpenEBSSnapshot == nil {
		return fmt.Errorf("invalid VolumeSnapshotDataSource: %v", src)
	}

	casType := pv.Annotations["openebs.io/cas-type"]
	if casType == "" {
		casType = "jiva"
	}

	_, err := h.DeleteSnapshot(casType, pv.Name, src.OpenEBSSnapshot.SnapshotID, pv.Spec.ClaimRef.Namespace)
	if err != nil {
		glog.Errorf("failed to create snapshot for volume :%v, err: %v", pv.Name, err)
		return err
	}

	glog.V(1).Infof("snapshot deleted :%v successfully", src.OpenEBSSnapshot.SnapshotID)
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
	return nil, nil, fmt.Errorf("Snapshot not found")
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
	var newVolume v1alpha1.CASVolume
	var openebsVol mvol_v1alpha1.CASVolume

	volumeSpec, class := CreateCloneVolumeSpec(snapshotData, pvc, pvName)

	err := openebsVol.CreateVolume(volumeSpec)
	if err != nil {
		glog.Errorf("Error creating volume: %v", err)
		return nil, nil, err
	}
	err = openebsVol.ReadVolume(pvName, pvc.Namespace, class, &newVolume)
	if err != nil {
		glog.Errorf("Error getting volume details: %v", err)
		return nil, nil, err
	}

	if err != nil {
		glog.Errorf("snapshot :%v restore failed, err:%v", snapshotData.Spec.OpenEBSSnapshot.SnapshotID, err)
		return nil, nil, fmt.Errorf("failed to restore %s, err: %v", snapshotData.Spec.OpenEBSSnapshot.SnapshotID, err)
	}

	glog.V(1).Infof("snapshot restored successfully to: %v", snapshotData.Spec.OpenEBSSnapshot.SnapshotID)

	vollabels := make(map[string]string)
	vollabels = provisioner.Setlink(vollabels, pvName)
	vollabels[string(v1alpha1.CASTypeKey)] = newVolume.Spec.CasType
	vollabels[string(v1alpha1.StorageClassKey)] = class

	pv := &v1.PersistentVolumeSource{
		ISCSI: &v1.ISCSIPersistentVolumeSource{
			TargetPortal: newVolume.Spec.TargetPortal,
			IQN:          newVolume.Spec.Iqn,
			Lun:          0,
			FSType:       "ext4",
			ReadOnly:     false,
		},
	}
	return pv, vollabels, nil
}

// VolumeDelete deletes the persistent volume
func (h *openEBSPlugin) VolumeDelete(pv *v1.PersistentVolume) error {
	if pv == nil || pv.Spec.ISCSI == nil {
		return fmt.Errorf("invalid VolumeSnapshotDataSource: %v", pv)
	}
	var openebsVol mvol_v1alpha1.CASVolume

	err := openebsVol.DeleteVolume(pv.Name, pv.Spec.ClaimRef.Namespace)
	if err != nil {
		glog.Errorf("Error while deleting volume: %v", err)
		return err
	}
	return nil
}

func GetMayaService() error {
	client, err := GetK8sClient()
	if err != nil {
		return err
	}
	var openebsVol mvol_v1alpha1.CASVolume

	//Get maya-apiserver IP address from cluster
	addr, err := openebsVol.GetMayaClusterIP(client)
	if err != nil {
		glog.Errorf("Error getting maya-apiserver IP Address: %v", err)
		return err
	}
	mayaServiceURI := "http://" + addr + ":5656"
	//Set maya-apiserver IP address along with default port
	os.Setenv("MAPI_ADDR", mayaServiceURI)

	return nil
}

// GetPersistentVolumeClass returns StorageClassName
func GetPersistentVolumeClass(volume *v1.PersistentVolume) string {
	// Use label first
	if class, found := volume.Labels[string(v1alpha1.StorageClassKey)]; found {
		return class
	}

	return volume.Spec.StorageClassName
}

// GetPersistentClass returns StoragClassName
func GetStorageClass(pvName string) (string, error) {
	client, err := GetK8sClient()
	if err != nil {
		return "", err
	}
	volume, err := client.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	glog.Infof("Source Volume is %#v", volume)
	return GetPersistentVolumeClass(volume), nil
}

// GetNameAndNameSpaceFromSnapshotName retrieves the namespace and
// the short name of a snapshot from its full name, for exmaple
// "test-ns/snap1"
func GetNameAndNameSpaceFromSnapshotName(name string) (string, string, error) {
	strs := strings.Split(name, "/")
	if len(strs) != 2 {
		return "", "", fmt.Errorf("invalid snapshot name")
	}
	return strs[0], strs[1], nil
}

// CreateVolumeSpec constructs the volumeSpec for volume create request
func CreateCloneVolumeSpec(snapshotData *crdv1.VolumeSnapshotData,
	pvc *v1.PersistentVolumeClaim,
	pvName string,
) (vol v1alpha1.CASVolume,
	class string) {

	// restore snapshot to a PV
	// get the snaphot ID and source volume
	snapshotID := snapshotData.Spec.OpenEBSSnapshot.SnapshotID
	pvRefName := snapshotData.Spec.PersistentVolumeRef.Name
	//pvRefNamespace := snapshotData.Spec.PersistentVolumeRef.Namespace
	casVolume := v1alpha1.CASVolume{}

	mapLabels := make(map[string]string)

	// Get the source PV storage class name which will be passed
	// to maya-apiserver to extract volume policy while restoring snapshot as
	// new volume.
	pvRefStorageClass, err := GetStorageClass(pvRefName)
	if err != nil {
		glog.Errorf("Error getting volume details: %v", err)
	}
	if len(pvRefStorageClass) == 0 {
		glog.Errorf("Volume has no storage class specified")
	} else {

		mapLabels[string("openebs.io/storageclass")] = pvRefStorageClass
		casVolume.Labels = mapLabels
	}
	glog.Infof("Using the Storage Class %s for dynamic provisioning", pvRefStorageClass)

	// construct volumespec for volume create request.
	// Enable volume clone: set clone as true, enables openebs volume to be created
	// as a clone volume
	volSize := pvc.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	casVolume.Spec.Capacity = volSize.String()
	casVolume.Namespace = pvc.Namespace
	casVolume.Labels[string(v1alpha1.NamespaceKey)] = pvc.Namespace
	casVolume.Labels[string(v1alpha1.PersistentVolumeClaimKey)] = pvc.ObjectMeta.Name
	casVolume.ObjectMeta.Name = pvName
	casVolume.CloneSpec.SnapshotName = snapshotID
	casVolume.CloneSpec.IsClone = true
	casVolume.CloneSpec.SourceVolume = pvRefName

	return casVolume, pvRefStorageClass
}
