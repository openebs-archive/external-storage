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

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"syscall"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/kubernetes-incubator/external-storage/lib/util"
	"github.com/kubernetes-incubator/external-storage/openebs/pkg/tracing"
	mApiv1 "github.com/kubernetes-incubator/external-storage/openebs/pkg/v1"
	mayav1 "github.com/kubernetes-incubator/external-storage/openebs/types/v1"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	provisionerName = "openebs.io/provisioner-iscsi"
	// BetaStorageClassAnnotation represents the beta/previous StorageClass annotation.
	// It's currently still used and will be held for backwards compatibility
	BetaStorageClassAnnotation = "volume.beta.kubernetes.io/storage-class"
)

type openEBSProvisioner struct {
	// Maya-API Server URI running in the cluster
	mapiURI string

	// Identity of this openEBSProvisioner, set to node's name. Used to identify
	// "this" provisioner's PVs.
	identity string

	tracer opentracing.Tracer

	closer io.Closer
}

// NewOpenEBSProvisioner creates a new openebs provisioner
func NewOpenEBSProvisioner(client kubernetes.Interface) controller.Provisioner {
	tracer := tracing.Init("provisioner")
	opentracing.SetGlobalTracer(tracer)
	span := tracer.StartSpan("Return OpenEBS Provisioner instance")

	ctx := opentracing.ContextWithSpan(context.Background(), span)
	defer span.Finish()
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		span.LogFields(
			log.String("event", "Getenv"),
			log.Error(fmt.Errorf("ENV variable NODE_NAME not set NODE_NAME : %v", nodeName)),
		)
		glog.Errorf("ENV variable 'NODE_NAME' is not set")
	}
	var openebsObj mApiv1.OpenEBSVolume
	span.LogFields(
		log.String("event", "get node name"),
		log.String("NODE_NAME", nodeName),
	)
	//Get maya-apiserver IP address from cluster
	addr, err := openebsObj.GetMayaClusterIP(ctx, client)

	if err != nil {
		span.LogFields(
			log.String("event", "get maya cluster IP"),
			log.Error(err),
		)
		glog.Errorf("Error getting maya-apiserver IP Address: %v", err)
		return nil
	}

	mayaServiceURI := "http://" + addr + ":5656"
	span.LogFields(
		log.String("event", "get maya service URI"),
		log.String("mayaServiceURI", mayaServiceURI),
	)

	//Set maya-apiserver IP address along with default port
	os.Setenv("MAPI_ADDR", mayaServiceURI)
	return &openEBSProvisioner{
		mapiURI:  mayaServiceURI,
		identity: nodeName,
		tracer:   tracer,
	}
}

var _ controller.Provisioner = &openEBSProvisioner{}

// Provision creates a storage asset and returns a PV object representing it.
func (p *openEBSProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {

	span := p.tracer.StartSpan("req provision volume")
	span.SetTag("volume-name", options.PVName)
	span.SetBaggageItem("operation", "createVolume")
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)

	//Issue a request to Maya API Server to create a volume
	var volume mayav1.Volume
	var openebsVol mApiv1.OpenEBSVolume
	volumeSpec := mayav1.VolumeSpec{}

	volSize := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	volumeSpec.Metadata.Labels.Storage = volSize.String()

	className := GetStorageClassName(ctx, options)

	if className == nil {
		span.LogFields(
			log.String("event", "get storage class name"),
			log.Error(fmt.Errorf("could not get class name, classname : %v", className)),
		)
		glog.Errorf("Volume has no storage class specified")
	} else {
		volumeSpec.Metadata.Labels.StorageClass = *className
	}
	volumeSpec.Metadata.Labels.Namespace = options.PVC.Namespace
	volumeSpec.Metadata.Name = options.PVName
	_, err := openebsVol.CreateVolume(ctx, volumeSpec)
	if err != nil {
		span.LogFields(
			log.String("event", "create volume"),
			log.Bool("success", false),
			log.Error(err),
		)
		glog.Errorf("Error creating volume: %v", err)
		return nil, err
	}

	span.LogFields(
		log.String("event", "create volume"),
		log.Bool("success", true),
	)

	span.SetBaggageItem("operation", "listVolume")
	err = openebsVol.ListVolume(ctx, options.PVName, options.PVC.Namespace, &volume)
	if err != nil {
		span.LogFields(
			log.String("event", "list volume"),
			log.Error(err),
		)
		glog.Errorf("Error getting volume details: %v", err)
		return nil, err
	}

	span.LogFields(
		log.String("event", "list volume"),
		log.Bool("success", true),
	)
	// Use annotations to specify the context using which the PV was created.
	volAnnotations := make(map[string]string)
	volAnnotations["openEBSProvisionerIdentity"] = p.identity

	var iqn, targetPortal string

	for key, value := range volume.Metadata.Annotations.(map[string]interface{}) {
		switch key {
		case "vsm.openebs.io/iqn":
			iqn = value.(string)
		case "vsm.openebs.io/targetportals":
			targetPortal = value.(string)
		}
	}

	span.LogFields(
		log.Object("iqn", iqn),
		log.Object("target", targetPortal),
	)

	glog.V(2).Infof("Volume IQN: %v , Volume Target: %v", iqn, targetPortal)

	if !util.AccessModesContainedInAll(p.GetAccessModes(), options.PVC.Spec.AccessModes) {
		span.LogFields(
			log.String("event", "get access modes"),
			log.Object("access-mode", options.PVC.Spec.AccessModes),
			log.Object("supported-mode", p.GetAccessModes()),
			log.Error(fmt.Errorf("access mode not supported")),
		)
		glog.V(1).Info("Invalid Access Modes: %v, Supported Access Modes: %v", options.PVC.Spec.AccessModes, p.GetAccessModes())
		return nil, fmt.Errorf("Invalid Access Modes: %v, Supported Access Modes: %v", options.PVC.Spec.AccessModes, p.GetAccessModes())
	}
	span.LogFields(
		log.String("event", "get access modes"),
		log.Object("access-mode", options.PVC.Spec.AccessModes),
		log.Object("supported-mode", p.GetAccessModes()),
	)
	// The following will be used by the dashboard, to display links on PV page
	userLinks := make([]string, 0)
	localMonitoringURL := os.Getenv("OPENEBS_MONITOR_URL")
	if localMonitoringURL != "" {
		localMonitorLinkName := os.Getenv("OPENEBS_MONITOR_LINK_NAME")
		if localMonitorLinkName == "" {
			localMonitorLinkName = "monitor"
		}
		localMonitorVolKey := os.Getenv("OPENEBS_MONITOR_VOLKEY")
		if localMonitorVolKey != "" {
			localMonitoringURL += localMonitorVolKey + "=" + options.PVName
		}
		userLinks = append(userLinks, "\""+localMonitorLinkName+"\":\""+localMonitoringURL+"\"")
	}

	span.LogFields(
		log.String("event", "get monitoring URL"),
		log.Object("URL", userLinks),
	)
	mayaPortalURL := os.Getenv("MAYA_PORTAL_URL")
	if mayaPortalURL != "" {
		mayaPortalLinkName := os.Getenv("MAYA_PORTAL_LINK_NAME")
		if mayaPortalLinkName == "" {
			mayaPortalLinkName = "maya"
		}
		userLinks = append(userLinks, "\""+mayaPortalLinkName+"\":\""+mayaPortalURL+"\"")
	}
	span.LogFields(
		log.String("event", "get maya-portal URL"),
		log.Object("URL", mayaPortalURL),
	)
	if len(userLinks) > 0 {
		volAnnotations["alpha.dashboard.kubernetes.io/links"] = "{" + strings.Join(userLinks, ",") + "}"
	}

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.PVName,
			Annotations: volAnnotations,
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
					Lun:          0,
					FSType:       "ext4",
					ReadOnly:     false,
				},
			},
		},
	}

	span.LogFields(
		log.String("event", "volume-provisioned-successfully"),
		log.Object("pv-details", pv),
	)
	return pv, nil
}

// Delete removes the storage asset that was created by Provision represented
// by the given PV.
func (p *openEBSProvisioner) Delete(volume *v1.PersistentVolume) error {

	span := p.tracer.StartSpan("req delete volume")
	span.SetTag("volume-details", volume)
	span.SetBaggageItem("operation", "deleteVolume")
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)

	var openebsVol mApiv1.OpenEBSVolume

	ann, ok := volume.Annotations["openEBSProvisionerIdentity"]
	if !ok {
		span.LogFields(
			log.String("event", "get volume annotations"),
			log.Bool("success", false),
			log.Error(fmt.Errorf("identity annotation not found on PV")),
		)
		return errors.New("identity annotation not found on PV")
	}
	if ann != p.identity {
		span.LogFields(
			log.String("event", "get volume identity"),
			log.Bool("success", false),
			log.Error(fmt.Errorf("identity annotation on PV does not match with kubernetes controller")),
		)
		return &controller.IgnoredError{Reason: "identity annotation on PV does not match ours"}
	}

	// Issue a delete request to Maya API Server
	err := openebsVol.DeleteVolume(ctx, volume.Name, volume.Spec.ClaimRef.Namespace)
	if err != nil {
		span.LogFields(
			log.String("event", "delete volume"),
			log.String("volume name", volume.Name),
			log.Bool("success", false),
			log.Error(err),
		)

		glog.Errorf("Error while deleting volume: %v", err)
		return err
	}
	span.LogFields(
		log.String("event", "delete volume"),
		log.String("volume name", volume.Name),
		log.Bool("success", true),
	)
	return nil
}

func (p *openEBSProvisioner) GetAccessModes() []v1.PersistentVolumeAccessMode {
	return []v1.PersistentVolumeAccessMode{
		v1.ReadWriteOnce,
	}
}

func main() {
	syscall.Umask(0)

	flag.Parse()
	flag.Set("logtostderr", "true")

	var (
		config     *rest.Config
		err        error
		k8sMaster  = mayav1.K8sMasterENV()
		kubeConfig = mayav1.KubeConfigENV()
	)
	if len(k8sMaster) != 0 || len(kubeConfig) != 0 {
		fmt.Printf("Build client config using k8s Master's Address: '%s' or Kubeconfig: '%s' \n", k8sMaster, kubeConfig)
		config, err = clientcmd.BuildConfigFromFlags(k8sMaster, kubeConfig)
	} else {
		// Create an InClusterConfig and use it to create a client for the controller
		// to use to communicate with Kubernetes
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		glog.Errorf("Failed to create config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorf("Failed to create client: %v", err)
	}
	// The controller needs to know what the server version is because out-of-tree
	// provisioners aren't officially supported until 1.5
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		glog.Errorf("Error getting server version: %v", err)
	}
	// Create the provisioner: it implements the Provisioner interface expected by
	// the controller
	openEBSProvisioner := NewOpenEBSProvisioner(clientset)
	if openEBSProvisioner != nil {
		// Start the provision controller which will dynamically provision OpenEBS VSM
		// PVs
		pc := controller.NewProvisionController(
			clientset,
			provisionerName,
			openEBSProvisioner,
			serverVersion.GitVersion,
		)
		pc.Run(wait.NeverStop)
	} else {
		os.Exit(1) //Exit if provisioner not created.
	}

}

// GetStorageClassName returns StorageClassName.
func GetStorageClassName(ctx context.Context, options controller.VolumeOptions) *string {

	span, _ := opentracing.StartSpanFromContext(ctx, "get-storage-class-name")
	defer span.Finish()
	// Use beta annotation first
	if class, found := options.PVC.Annotations[BetaStorageClassAnnotation]; found {
		span.LogFields(
			log.String("event", "get storage class"),
			log.Object("storage-class", class),
		)
		return &class
	}
	span.LogFields(
		log.String("event", "get storage class"),
		log.Object("storage-class", options.PVC.Spec.StorageClassName),
	)
	return options.PVC.Spec.StorageClassName
}
