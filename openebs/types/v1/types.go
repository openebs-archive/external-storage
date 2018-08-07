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

package v1

const (
	PVCLabelsApplication = "volumeprovisioner.mapi.openebs.io/application"
	PVCLabelsReplicaTopKeyDomain = "volumeprovisioner.mapi.openebs.io/replica-topology-key-domain"
	PVCLabelsReplicaTopKeyType = "volumeprovisioner.mapi.openebs.io/replica-topology-key-type"
)

//VolumeSpec holds the config for creating a OpenEBS Volume
type VolumeSpec struct {
	Kind       string `yaml:"kind"`
	APIVersion string `yaml:"apiVersion"`
	Metadata   struct {
		Name   string `yaml:"name"`
		Labels struct {
			Storage               string `yaml:"volumeprovisioner.mapi.openebs.io/storage-size"`
			StorageClass          string `yaml:"k8s.io/storage-class"`
			Namespace             string `yaml:"k8s.io/namespace"`
			PersistentVolumeClaim string `yaml:"k8s.io/pvc"`
			Application           string `yaml:"volumeprovisioner.mapi.openebs.io/application,omitempty"`
			ReplicaTopoKeyDomain  string `yaml:"volumeprovisioner.mapi.openebs.io/replica-topology-key-domain,omitempty"`
			ReplicaTopoKeyType    string `yaml:"volumeprovisioner.mapi.openebs.io/replica-topology-key-type,omitempty"`
		} `yaml:"labels"`
	} `yaml:"metadata"`
	VolumeClone `yaml:"volumeClone"`
}

type VolumeClone struct {
	// Defaults to false, true will enable the volume to be created as a clone
	Clone bool `yaml:"clone,omitempty"`
	// SourceVolume is snapshotted volume
	SourceVolume string `yaml:"sourceVolume,omitempty"`
	// CloneIP is the source controller IP which will be used to make a sync and rebuild
	// request from the new clone replica.
	CloneIP string `yaml:"cloneIP,omitempty"`
	// SnapshotName name of snapshot which is getting promoted as persistent
	// volume(this snapshot will be cloned to new volume).
	SnapshotName string `yaml:"snapshotName,omitempty"`
}

// Volume is a command implementation struct
type Volume struct {
	Spec struct {
		AccessModes interface{} `json:"AccessModes"`
		Capacity    interface{} `json:"Capacity"`
		ClaimRef    interface{} `json:"ClaimRef"`
		OpenEBS     struct {
			VolumeID string `json:"volumeID"`
		} `json:"OpenEBS"`
		PersistentVolumeReclaimPolicy string `json:"PersistentVolumeReclaimPolicy"`
		StorageClassName              string `json:"StorageClassName"`
	} `json:"Spec"`

	Status struct {
		Message string `json:"Message"`
		Phase   string `json:"Phase"`
		Reason  string `json:"Reason"`
	} `json:"Status"`
	Metadata struct {
		Annotations       interface{} `json:"annotations"`
		CreationTimestamp interface{} `json:"creationTimestamp"`
		Name              string      `json:"name"`
	} `json:"metadata"`
}

// SnapshotAPISpec hsolds the config for creating asnapshot of volume
type SnapshotAPISpec struct {
	Kind       string `yaml:"kind"`
	APIVersion string `yaml:"apiVersion"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		VolumeName string `yaml:"volumeName"`
	} `yaml:"spec"`
}
