package openebs

import (
	"errors"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ErrK8SApiAccountNotSet is returned when the account used to talk to k8s api
// is not setup
var ErrK8SApiAccountNotSet = errors.New("k8s api service-account is not setup")

// GetK8sClient instantiates a k8s client
func GetK8sClient() (*kubernetes.Clientset, error) {
	var k8sClient *kubernetes.Clientset
	var err error
	kubeconfig := os.Getenv("KUBECONFIG")
	if len(kubeconfig) > 0 {
		k8sClient, err = loadClientFromKubeconfig(kubeconfig)
	} else {
		k8sClient, err = loadClientFromServiceAccount()
	}
	if err != nil {
		return nil, err
	}
	if k8sClient == nil {
		return nil, ErrK8SApiAccountNotSet
	}
	return k8sClient, nil
}

// loadClientFromServiceAccount loads a k8s client from a ServiceAccount
// specified in the pod running
func loadClientFromServiceAccount() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
}

// loadClientFromKubeconfig loads a k8s client from a kubeconfig flag
func loadClientFromKubeconfig(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil

}

// GetPersistentVolumeClass returns StorageClassName.
/*func GetPersistentVolumeClass(volume *v1.PersistentVolume) string {
	// Use beta annotation first
	if class, found := volume.Annotations[core.BetaStorageClassAnnotation]; found {
		return class
	}

	return volume.Spec.StorageClassName
}
*/
