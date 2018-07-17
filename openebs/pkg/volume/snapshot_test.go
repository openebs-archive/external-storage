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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	utiltesting "k8s.io/client-go/util/testing"
)

var (
	snapshotResponse      = `{"actions":{},"id":"snap1","links":{"self":"http://10.36.0.1:9501/v1/snapshotoutputs/snap1"},"type":"snapshotOutput"}`
	volumeNameIsMissing   = errors.New("Volume name is missing")
	snapshotNameIsMissing = errors.New("Snapshot name is missing")
	badReqErr             = errors.New(snapshotResponse)
	volNotFound           = errors.New("Volume not found")
	MAPIADDRNotSet        = errors.New("MAPI_ADDR environment variable not set")
)

func TestCreateSnapshot(t *testing.T) {
	tests := map[string]struct {
		volumeName  string
		snapName    string
		fakeHandler utiltesting.FakeHandler
		err         error
		addr        string
		expectResp  string
	}{
		"StatusOK": {
			volumeName: "testvol",
			snapName:   "snap1",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(snapshotResponse),
				T:            t,
			},
			err:        nil,
			addr:       "MAPI_ADDR",
			expectResp: "Snapshot Successfully Created",
		},
		"BadRequest": {
			volumeName: "12324rty653423",
			snapName:   "134efvet454",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: string(snapshotResponse),
				T:            t,
			},
			err:        badReqErr,
			addr:       "MAPI_ADDR",
			expectResp: "HTTP Status error from maya-apiserver",
		},
		"MAPI_ADDRNotSet": {
			volumeName: "testvol",
			snapName:   "snap1",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(snapshotResponse),
			},
			err:        MAPIADDRNotSet,
			addr:       "",
			expectResp: "Error getting maya-apiserver IP Address",
		},
		"VolumeNameMissing": {
			volumeName: "",
			snapName:   "snap1",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume name is missing",
				T:            t,
			},
			err:        volumeNameIsMissing,
			addr:       "MAPI_ADDR",
			expectResp: "HTTP Status error from maya-apiserver",
		},
		"VolumeNotFound": {
			volumeName: "test12345",
			snapName:   "snap1",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume not found",
				T:            t,
			},
			err:        volNotFound,
			addr:       "MAPI_ADDR",
			expectResp: "HTTP Status error from maya-apiserver",
		},
		"SnapshotNameMissing": {
			volumeName: "testvol",
			snapName:   "",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Snapshot name is missing",
				T:            t,
			},
			err:        snapshotNameIsMissing,
			addr:       "MAPI_ADDR",
			expectResp: "HTTP Status error from maya-apiserver",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			var vol OpenEBSVolume
			resp, err := vol.CreateSnapshot(tt.volumeName, tt.snapName, "default")
			if !reflect.DeepEqual(err, tt.err) {
				t.Fatalf("CreateSnapshot(%v, %v) => got %v, want %v ", tt.volumeName, tt.snapName, err, tt.err)
			}
			if !reflect.DeepEqual(resp, tt.expectResp) {
				t.Fatalf("CreateSnapshot(%v, %v) => got %v, want %v ", tt.volumeName, tt.snapName, resp, tt.expectResp)
			}
		})
	}
}

func TestRevertSnapshot(t *testing.T) {
	tests := map[string]struct {
		volumeName  string
		snapName    string
		fakeHandler utiltesting.FakeHandler
		err         error
		addr        string
	}{
		"StatusOK": {
			volumeName: "testvol",
			snapName:   "snap1",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(snapshotResponse),
				T:            t,
			},
			err:  nil,
			addr: "MAPI_ADDR",
		},
		"BadRequest": {
			volumeName: "testvol",
			snapName:   "snap1",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Server status error: Bad Request",
				T:            t,
			},
			err:  fmt.Errorf("Server status error: %v", http.StatusText(400)),
			addr: "MAPI_ADDR",
		},
		"MAPI_ADDRNotSet": {
			volumeName: "testvol",
			snapName:   "snap1",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(snapshotResponse),
			},
			err:  MAPIADDRNotSet,
			addr: "",
		},
		"VolumeNameMissing": {
			volumeName: "",
			snapName:   "snap1",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume name is missing",
				T:            t,
			},
			err:  volumeNameIsMissing,
			addr: "MAPI_ADDR",
		},
		"VolumeNotFound": {
			volumeName: "sbc12234",
			snapName:   "snap_123143545",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume not found",
				T:            t,
			},
			err:  volNotFound,
			addr: "MAPI_ADDR",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			var vol OpenEBSVolume
			_, err := vol.RevertSnapshot(tt.volumeName, tt.snapName)
			if !reflect.DeepEqual(err, tt.err) {
				t.Fatalf("RevertSnapshot(%v, %v) => got %v, want %v ", tt.volumeName, tt.snapName, err, tt.err)
			}
		})
	}
}

func TestListSnapshot(t *testing.T) {
	tests := map[string]struct {
		volumeName  string
		fakeHandler utiltesting.FakeHandler
		err         error
		addr        string
	}{
		"StatusOK": {
			volumeName: "testvol",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: "snap1",
				T:            t,
			},
			err:  fmt.Errorf("EOF"),
			addr: "MAPI_ADDR",
		},
		"BadRequest": {
			volumeName: "12324rty653423",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Server status error: Bad Request",
				T:            t,
			},
			err:  fmt.Errorf("Server status error: %v", http.StatusText(400)),
			addr: "MAPI_ADDR",
		},
		"MAPI_ADDRNotSet": {
			volumeName: "testvol",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(snapshotResponse),
			},
			err:  MAPIADDRNotSet,
			addr: "",
		},
		"VolumeNameMissing": {
			volumeName: "",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume name is missing",
				T:            t,
			},
			err:  volumeNameIsMissing,
			addr: "MAPI_ADDR",
		},
		"VolumeNotFound": {
			volumeName: "abc1234",
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   400,
				ResponseBody: "Volume not found",
				T:            t,
			},
			err:  volNotFound,
			addr: "MAPI_ADDR",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			os.Setenv(tt.addr, server.URL)
			defer os.Unsetenv(tt.addr)
			defer server.Close()
			var obj interface{}
			var vol OpenEBSVolume
			err := vol.ListSnapshot(tt.volumeName, "", "default", obj)
			if !reflect.DeepEqual(err, tt.err) {
				t.Fatalf("ListSnapshot(%v) => got %v, want %v ", tt.volumeName, err, tt.err)
			}
		})
	}
}
