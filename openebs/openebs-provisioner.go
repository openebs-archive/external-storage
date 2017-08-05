/*
Copyright 2016-2017 The Kubernetes Authors.
Copyright 2016-2017 The OpenEBS Authors.

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

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"syscall"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	timeout                   = 60 * time.Second
	resyncPeriod              = 15 * time.Second
	provisionerName           = "openebs.io/provisioner-iscsi"
	exponentialBackOffOnError = false
	failedRetryThreshold      = 5
	leasePeriod               = controller.DefaultLeaseDuration
	retryPeriod               = controller.DefaultRetryPeriod
	renewDeadline             = controller.DefaultRenewDeadline
	termLimit                 = controller.DefaultTermLimit
)

//VsmSpec holds the config for creating a VSM
type VsmSpec struct {
	Kind       string `yaml:"kind"`
	APIVersion string `yaml:"apiVersion"`
	Metadata   struct {
		Name   string `yaml:"name"`
		Labels struct {
			Storage string `yaml:"volumeprovisioner.mapi.openebs.io/storage-size"`
		}
	} `yaml:"metadata"`
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
		glog.Fatal("env variable NODE_NAME must be set so that this provisioner can identify itself")
	}

	//TODO - HandleError Cases
	mayaServiceURI := "http://" + getMayaClusterIP(client) + ":5656"
	os.Setenv("MAPI_ADDR", mayaServiceURI)

	return &openEBSProvisioner{
		mapiURI:  mayaServiceURI,
		identity: nodeName,
	}
}

var _ controller.Provisioner = &openEBSProvisioner{}

// Provision creates a storage asset and returns a PV object representing it.
func (p *openEBSProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	//path := "/var/openebs/" + options.PVName

	//TODO - Issue a request to Maya API Server to create a volume
	var volume Volume
	volSize := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	//TODO - Need to change the size as a value that Maya Server can accept
	err := createVsm(options.PVName, volSize.String())
	if err != nil {
		glog.Fatalf("Error creating volume: %v", err)
		return nil, err
	}

	err = listVsm(options.PVName, &volume)
	if err != nil {
		glog.Fatalf("Error getting volume details: %v", err)
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

	glog.Infof("Volume IQN: %v", iqn)
	glog.Infof("Volume Target: %v", targetPortal)

	//TODO - fill in the iSCSI PV details
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.PVName,
			Annotations: map[string]string{
				"openEBSProvisionerIdentity": p.identity,
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				ISCSI: &v1.ISCSIVolumeSource{
					TargetPortal: targetPortal,
					IQN:          iqn,
					Lun:          1,
					FSType:       "ext4",
					ReadOnly:     false,
				},
			},
		},
	}

	return pv, nil
}

// Delete removes the storage asset that was created by Provision represented
// by the given PV.
func (p *openEBSProvisioner) Delete(volume *v1.PersistentVolume) error {
	ann, ok := volume.Annotations["openEBSProvisionerIdentity"]
	if !ok {
		return errors.New("identity annotation not found on PV")
	}
	if ann != p.identity {
		return &controller.IgnoredError{Reason: "identity annotation on PV does not match ours"}
	}

	//TODO - Issue a delete request to Maya API Server
	deleteVsm(volume.Name)

	return nil
}

func getMayaClusterIP(client kubernetes.Interface) string {
	clusterIP := "127.0.0.1"

	//Fetch the Maya ClusterIP using the Maya API Server Service
	sc, err := client.CoreV1().Services("default").Get("maya-apiserver-service", metav1.GetOptions{})
	if err != nil {
		glog.Fatalf("Error getting maya-api-server IP Address: %v", err)
	}

	clusterIP = sc.Spec.ClusterIP
	glog.Infof("Maya Cluster IP: %v", clusterIP)

	return clusterIP
}

// createVsm to create the Vsm through a API call to m-apiserver
func createVsm(vname string, size string) error {

	var vs VsmSpec

	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		glog.Fatalf("Error getting maya-api-server IP Address: %v", err)
		return err
	}
	url := addr + "/latest/volumes/"

	vs.Kind = "PersistentVolumeClaim"
	vs.APIVersion = "v1"
	vs.Metadata.Name = vname
	vs.Metadata.Labels.Storage = size

	//Marshal serializes the value provided into a YAML document
	yamlValue, _ := yaml.Marshal(vs)

	glog.Infof("VSM Spec Created:\n%v\n", string(yamlValue))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(yamlValue))

	req.Header.Add("Content-Type", "application/yaml")

	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		glog.Fatalf("http.Do() error: : %v", err)
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Fatalf("ioutil.ReadAll() error: : %v", err)
		return err
	}

	code := resp.StatusCode
	if code != http.StatusOK {
		glog.Fatalf("Status error: %v\n", http.StatusText(code))
		os.Exit(1)
	}
	glog.Infof("VSM Successfully Created:\n%v\n", string(data))
	return nil
}

// listVsm to get the info of Vsm through a API call to m-apiserver
func listVsm(vname string, obj interface{}) error {

	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		glog.Fatalf("Error getting maya-api-server IP Address: %v", err)
		return err
	}
	url := addr + "/latest/volumes/info/" + vname

	glog.Infof("Get details for VSM :%v", string(vname))

	req, err := http.NewRequest("GET", url, nil)
	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		glog.Fatalf("http.Do() error: : %v", err)
		return err
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code != http.StatusOK {
		glog.Fatalf("Status error: %v\n", http.StatusText(code))
		os.Exit(1)
	}
	glog.Info("VSM Details Successfully Retrieved")
	return json.NewDecoder(resp.Body).Decode(obj)
}

// deleteVsm to get delete Vsm through a API call to m-apiserver
func deleteVsm(vname string) error {

	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		glog.Fatalf("Error getting maya-api-server IP Address: %v", err)
		return err
	}
	url := addr + "/latest/volumes/delete/" + vname

	glog.Infof("Delete VSM :%v", string(vname))

	req, err := http.NewRequest("GET", url, nil)
	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		glog.Fatalf("http.Do() error: : %v", err)
		return err
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code != http.StatusOK {
		glog.Fatalf("Status error: %v\n", http.StatusText(code))
		os.Exit(1)
	}
	glog.Info("VSM Deleted Successfully initiated")
	return nil
}

func main() {
	syscall.Umask(0)

	flag.Parse()
	flag.Set("logtostderr", "true")

	// Create an InClusterConfig and use it to create a client for the controller
	// to use to communicate with Kubernetes
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatalf("Failed to create config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Failed to create client: %v", err)
	}

	// The controller needs to know what the server version is because out-of-tree
	// provisioners aren't officially supported until 1.5
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		glog.Fatalf("Error getting server version: %v", err)
	}

	// Create the provisioner: it implements the Provisioner interface expected by
	// the controller
	openEBSProvisioner := NewOpenEBSProvisioner(clientset)

	// Start the provision controller which will dynamically provision OpenEBS VSM
	// PVs
	pc := controller.NewProvisionController(
		clientset,
		provisionerName,
		openEBSProvisioner,
		serverVersion.GitVersion)
		// resyncPeriod,
		// exponentialBackOffOnError,
		// failedRetryThreshold,
		// leasePeriod,
		// renewDeadline,
		// retryPeriod,
		// termLimit)
	pc.Run(wait.NeverStop)
}

/*
	client kubernetes.Interface,
	provisionerName string,
	provisioner Provisioner,
	kubeVersion string,
	options ...func(*ProvisionController) error,
*/
