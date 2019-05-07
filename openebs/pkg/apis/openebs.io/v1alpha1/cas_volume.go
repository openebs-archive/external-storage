/*
Copyright 2018 The OpenEBS Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CASKey is a typed string to represent CAS related annotations'
// or labels' keys
//
// Example 1 - Below is a sample CASVolume that makes use of some CASKey
// constants.
//
// NOTE:
//  This specification is sent by openebs provisioner as http create request in
// its payload.
//
// ```yaml
// kind: CASVolume
// apiVersion: v1alpha1
// metadata:
//   name: jiva-cas-vol
//   # this way of setting namespace gets the first priority
//   namespace: default
//   labels:
//     # this manner of setting namespace gets the second priority
//     openebs.io/namespace: default
//     openebs.io/pvc: openebs-repaffinity-0.6.0
// spec:
//   # latest way to set capacity
//   capacity: 2G
// ```
type CASKey string

const (
	// CASConfigKey is the key to fetch configurations w.r.t a CAS entity
	CASConfigKey CASKey = "cas.openebs.io/config"

	// NamespaceKey is the key to fetch cas entity's namespace
	NamespaceKey CASKey = "openebs.io/namespace"

	// PersistentVolumeClaimKey is the key to fetch name of PersistentVolumeClaim
	PersistentVolumeClaimKey CASKey = "openebs.io/persistentvolumeclaim"

	// IsPatchJivaReplicaNodeAffinityHeader is the key to fetch value of IsPatchKey
	// Its value is "Enable".
	IsPatchJivaReplicaNodeAffinityHeader CASKey = "Is-Patch-Jiva-Replica-Node-Affinity"

	// StorageClassKey is the key to fetch name of StorageClass
	StorageClassKey CASKey = "openebs.io/storageclass"

	// CASTypeKey is the key to fetch storage engine for the volume
	CASTypeKey CASKey = "openebs.io/cas-type"

	// StorageClassHeaderKey is the key to fetch name of StorageClass
	// This key is present only in get request headers
	StorageClassHeaderKey CASKey = "storageclass"
)

// CASVolumeKey is a typed string to represent CAS Volume related annotations'
// or labels' keys
//
// Example 1 - Below is a sample StorageClass that makes use of a CASVolumeKey
// constant i.e. the cas template used to create a cas volume
//
// ```yaml
// apiVersion: storage.k8s.io/v1
// kind: StorageClass
// metadata:
//  name: openebs-standard
//  annotations:
//    cas.openebs.io/create-volume-template: cast-standard-0.6.0
// provisioner: openebs.io/provisioner-iscsi
// ```
type CASVolumeKey string

const (
	// CASTemplateKeyForVolumeCreate is the key to fetch name of CASTemplate
	// to create a CAS Volume
	CASTemplateKeyForVolumeCreate CASVolumeKey = "cas.openebs.io/create-volume-template"

	// CASTemplateKeyForVolumeRead is the key to fetch name of CASTemplate
	// to read a CAS Volume
	CASTemplateKeyForVolumeRead CASVolumeKey = "cas.openebs.io/read-volume-template"

	// CASTemplateKeyForVolumeDelete is the key to fetch name of CASTemplate
	// to delete a CAS Volume
	CASTemplateKeyForVolumeDelete CASVolumeKey = "cas.openebs.io/delete-volume-template"

	// CASTemplateKeyForVolumeList is the key to fetch name of CASTemplate
	// to list CAS Volumes
	CASTemplateKeyForVolumeList CASVolumeKey = "cas.openebs.io/list-volume-template"
)

// CASVolume represents a cas volume
type CASVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec i.e. specifications of this cas volume
	Spec CASVolumeSpec `json:"spec"`
	// VolumeCloneSpec contains specifications required for volume clone
	CloneSpec VolumeCloneSpec `json:"cloneSpec,omitempty"`
	// Status of this cas volume
	Status CASVolumeStatus `json:"status"`
}

// CASVolumeSpec has the properties of a cas volume
type CASVolumeSpec struct {
	// Capacity will hold the capacity of this Volume
	Capacity string `json:"capacity"`
	// Iqn will hold the iqn value of this Volume
	Iqn string `json:"iqn"`
	// TargetPortal will hold the target portal for this volume
	TargetPortal string `json:"targetPortal"`
	// TargetIP will hold the targetIP for this Volume
	TargetIP string `json:"targetIP"`
	// TargetPort will hold the targetIP for this Volume
	TargetPort string `json:"targetPort"`
	// Replicas will hold the replica count for this volume
	Replicas string `json:"replicas"`
	// CasType will hold the storage engine used to provision this volume
	CasType string `json:"casType"`
	// FSType will hold the file system of the volume
	FSType string `json:"fsType"`
	// LUN will hold the lun of the volume
	Lun int32 `json:"lun"`
	// TODO add controller and replica status
}

// VolumeCloneSpec contains the required information which enable volume to cloned
type VolumeCloneSpec struct {
	// Defaults to false, true will enable the volume to be created as a clone
	IsClone bool `json:"isClone,omitempty"`
	// SourceVolume is snapshotted volume
	SourceVolume string `json:"sourceVolume,omitempty"`
	// CloneIP is the source controller IP which will be used to make a sync and rebuild
	// request from the new clone replica.
	SourceVolumeTargetIP string `json:"sourceTargetIP,omitempty"`
	// SnapshotName name of snapshot which is getting promoted as persistent
	// volume(this snapshot will be cloned to new volume).
	SnapshotName string `json:"snapshotName,omitempty"`
}

// CASVolumeStatus provides status of a cas volume
type CASVolumeStatus struct {
	// Phase indicates if a volume is available, pending or failed
	Phase VolumePhase
	// A human-readable message indicating details about why the volume
	// is in this state
	Message string
	// Reason is a brief CamelCase string that describes any failure and is meant
	// for machine parsing and tidy display in the CLI
	Reason string
}

// VolumePhase defines phase of a volume
type VolumePhase string

const (
	// VolumePending - used for Volumes that are not available
	VolumePending VolumePhase = "Pending"
	// VolumeAvailable - used for Volumes that are available
	VolumeAvailable VolumePhase = "Available"
	// VolumeFailed - used for Volumes that failed for some reason
	VolumeFailed VolumePhase = "Failed"
)

// CASVolumeList is a list of CASVolume resources
type CASVolumeList struct {
	metav1.ListOptions `json:",inline"`
	metav1.ObjectMeta  `json:"metadata,omitempty"`
	metav1.ListMeta    `json:"metalist"`

	// Items are the list of volumes
	Items []CASVolume `json:"items"`
}
