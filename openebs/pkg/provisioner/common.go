package provisioner

import (
	"fmt"
	"os"
	"strings"

	"github.com/kubernetes-incubator/external-storage/lib/controller"
)

const (
	// BetaStorageClassAnnotation represents the beta/previous StorageClass annotation.
	// It's currently still used and will be held for backwards compatibility
	BetaStorageClassAnnotation = "volume.beta.kubernetes.io/storage-class"
	//defaultFSType
	defaultFSType = "ext4"
)

// validFSType represents the valid fstype supported by openebs volume
// New supported type can be added using OPENEBS_VALID_FSTYPE env
var validFSType = []string{"ext4", "xfs"}

// GetPersistentVolumeClass returns StorageClassName.
func GetStorageClassName(options controller.VolumeOptions) *string {
	// Use beta annotation first
	if class, found := options.PVC.Annotations[BetaStorageClassAnnotation]; found {
		return &class
	}
	return options.PVC.Spec.StorageClassName
}

// parseClassParameters extract the new fstype other then "ext4"(dafault) which
// can be changed via "openebs.io/fstype" key and env OPENEBS_VALID_FSTYPE
func ParseClassParameters(params map[string]string) (string, error) {
	var fsType string
	for k, v := range params {
		switch strings.ToLower(k) {
		case "openebs.io/fstype":
			fsType = v
		}
	}
	if len(fsType) == 0 {
		fsType = defaultFSType
	}

	//Get openebs supported fstype from ENV variable
	validENVFSType := os.Getenv("OPENEBS_VALID_FSTYPE")
	if validENVFSType != "" {
		slices := strings.Split(validENVFSType, ",")
		for _, s := range slices {
			validFSType = append(validFSType, s)
		}
	}
	if !isValid(fsType, validFSType) {
		return "", fmt.Errorf("Filesystem %s is not supported", fsType)
	}
	return fsType, nil
}

// isValid checks the validity of fstype returns true if supported
func isValid(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
