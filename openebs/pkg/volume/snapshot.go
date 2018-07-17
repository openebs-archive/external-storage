/*
Copyright 2017 The OpenEBS Authors.

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

package volume

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/golang/glog"
	mayav1 "github.com/kubernetes-incubator/external-storage/openebs/types/v1"

	yaml "gopkg.in/yaml.v2"
)

// CreateSnapshot to create the Vsm through a API call to m-apiserver
func (v OpenEBSVolume) CreateSnapshot(volName string, snapName string, namespace string) (string, error) {
	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		return "Error getting maya-apiserver IP Address", err
	}

	var snap mayav1.SnapshotAPISpec

	snap.Kind = "VolumeSnapshot"
	snap.APIVersion = "v1"
	snap.Metadata.Name = snapName
	snap.Spec.VolumeName = volName

	url := addr + "/latest/snapshots/create/"

	//Marshal serializes the value provided into a YAML document
	yamlValue, _ := yaml.Marshal(snap)

	glog.V(2).Infof("snapshot Spec Created:\n%v\n", string(yamlValue))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(yamlValue))

	req.Header.Add("Content-Type", "application/yaml")

	req.Header.Set("namespace", namespace)
	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		glog.Errorf("Error when connecting maya-apiserver %v", err)
		return "Could not connect to maya-apiserver", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Unable to read response from maya-apiserver %v", err)
		return "Unable to read response from maya-apiserver", err
	}

	code := resp.StatusCode
	if err == nil && code != http.StatusOK {
		return "HTTP Status error from maya-apiserver", fmt.Errorf(string(data))
	}
	if code != http.StatusOK {
		glog.Errorf("Status error: %v\n", http.StatusText(code))
		return "HTTP Status error from maya-apiserver", err
	}

	glog.Infof("Snapshot Successfully Created:\n%v\n", string(data))
	return "Snapshot Successfully Created", nil
}

// ListVolume to get the info of Vsm through a API call to m-apiserver
func (v OpenEBSVolume) ListSnapshot(volName string, snapname string, namespace string, obj interface{}) error {

	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		return err
	}
	url := addr + "/latest/snapshots/list/"

	glog.V(2).Infof("[DEBUG] Get details for Volume :%v", string(volName))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("namespace", namespace)

	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		glog.Errorf("Error when connecting to maya-apiserver %v", err)
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Unable to read response from maya-apiserver %v", err)
		return err
	}
	code := resp.StatusCode
	if err == nil && code != http.StatusOK {
		return fmt.Errorf(string(data))
	}
	if code != http.StatusOK {
		glog.Errorf("HTTP Status error from maya-apiserver: %v\n", http.StatusText(code))
		return err
	}
	glog.V(2).Info("volume Details Successfully Retrieved")
	return json.NewDecoder(resp.Body).Decode(obj)
}

// RevertSnapshot revert a snapshot of volume by invoking the API call to m-apiserver
func (v OpenEBSVolume) RevertSnapshot(volName string, snapName string) (string, error) {
	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		return "Error getting maya-apiserver IP Address", err
	}

	var snap mayav1.SnapshotAPISpec
	snap.Metadata.Name = snapName
	snap.Spec.VolumeName = volName

	url := addr + "/latest/snapshots/revert/"

	yamlValue, _ := yaml.Marshal(snap)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(yamlValue))

	req.Header.Add("Content-Type", "application/yaml")

	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		glog.Errorf("Error when connecting maya-apiserver %v", err)
		return "Could not connect to maya-apiserver", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Unable to read response from maya-apiserver %v", err)
		return "Unable to read response from maya-apiserver", err
	}

	code := resp.StatusCode
	if err == nil && code != http.StatusOK {
		return "HTTP Status error from maya-apiserver", fmt.Errorf(string(data))
	}
	if code != http.StatusOK {
		glog.Errorf("Status error: %v\n", http.StatusText(code))
		return "HTTP Status error from maya-apiserver", err
	}

	glog.Infof("Snapshot Successfully restore:\n%v\n", string(data))
	return "Snapshot Successfully restore", nil

}

func (v OpenEBSVolume) SnapshotInfo(volName string, snapName string) (string, error) {

	return "Not implemented", nil
}

func (v OpenEBSVolume) DeleteSnapshot(snapName string) (string, error) {

	return "Not implemented", nil
}
