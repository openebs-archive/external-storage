package provisioner

import (
	"github.com/kubernetes-incubator/external-storage/lib/controller"
)

const (
	// BetaStorageClassAnnotation represents the beta/previous StorageClass annotation.
	// It's currently still used and will be held for backwards compatibility
	BetaStorageClassAnnotation = "volume.beta.kubernetes.io/storage-class"
)

// GetPersistentVolumeClass returns StorageClassName.
func GetStorageClassName(options controller.VolumeOptions) *string {
	// Use beta annotation first
	if class, found := options.PVC.Annotations[BetaStorageClassAnnotation]; found {
		return &class
	}
	return options.PVC.Spec.StorageClassName
}
