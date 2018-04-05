package v1

import (
	"os"
	"strings"
)

type ENVKey string

const (
	// KubeConfigENVK is the ENV key to fetch the kubeconfig
	KubeConfigENVK ENVKey = "OPENEBS_IO_KUBE_CONFIG"

	// K8sMasterENVK is the ENV key to fetch the K8s Master's Address
	K8sMasterENVK ENVKey = "OPENEBS_IO_K8S_MASTER"
)

func KubeConfigENV() string {
	val := GetEnv(KubeConfigENVK)
	return val
}

func K8sMasterENV() string {
	val := GetEnv(K8sMasterENVK)
	return val
}

// GetEnv fetches the environment variable value from the machine's
// environment
func GetEnv(envKey ENVKey) string {
	return strings.TrimSpace(os.Getenv(string(envKey)))
}
