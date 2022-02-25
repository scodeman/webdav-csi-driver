/*
Original code:
- https://github.com/cyverse/irods-csi-driver/blob/master/pkg/driver/driver.go

Following functions or objects are from the code under APL2 License.
- Driver
- NewDriver
- Run
- Stop
- isValidVolumeCapabilities
Original code:
- https://github.com/kubernetes-sigs/aws-fsx-csi-driver/blob/master/pkg/driver/driver.go
- https://github.com/kubernetes-sigs/aws-fsx-csi-driver/blob/master/pkg/driver/controller.go


Copyright 2019 The Kubernetes Authors.
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

package driver

import (
	"context"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"k8s.io/klog"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

const (
	driverName = "webdav.csi.scode.io"
)

var (
	volumeCaps = []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	}
)

// Driver object contains configuration parameters, grpc server and mounter
type Driver struct {
	config *Config

	server  *grpc.Server
	mounter Mounter
	secrets map[string]string

	nodeVolumes       map[string]*NodeVolume
	volumeLock        sync.Mutex
}

// NewDriver returns new driver
func NewDriver(conf *Config) *Driver {
	return &Driver{
		config:  conf,
		mounter: newNodeMounter(),
		secrets: make(map[string]string),

		nodeVolumes:       make(map[string]*NodeVolume),
		volumeLock:        sync.Mutex{},
	}
}

// Run runs the driver service
func (driver *Driver) Run() error {
	scheme, addr, err := ParseEndpoint(driver.config.Endpoint)
	if err != nil {
		return err
	}

	driver.secrets = make(map[string]string)
	secrets, err := ReadSecrets(driver.config.SecretPath)
	if err == nil {
		// if there's no secrets, it returns error, so we ignore
		// otherwise, copy
		for k, v := range secrets {
			driver.secrets[k] = v
		}
	}

	listener, err := net.Listen(scheme, addr)
	if err != nil {
		return err
	}

	logErr := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			klog.Errorf("GRPC error: %v", err)
		}
		return resp, err
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(logErr),
	}

	driver.server = grpc.NewServer(opts...)

	csi.RegisterIdentityServer(driver.server, driver)
	csi.RegisterControllerServer(driver.server, driver)
	csi.RegisterNodeServer(driver.server, driver)

	klog.Infof("Listening for connections on address: %#v", listener.Addr())
	return driver.server.Serve(listener)
}

// Stop stops the driver service
func (driver *Driver) Stop() {
	klog.Infof("Stopping server")
	driver.server.Stop()
}

// isValidVolumeCapabilities checks validity of volume capabilities
func (driver *Driver) isValidVolumeCapabilities(volCaps []*csi.VolumeCapability) bool {
	hasSupport := func(cap *csi.VolumeCapability) bool {
		for _, m := range volumeCaps {
			if m == cap.AccessMode.GetMode() {
				return true
			}
		}
		return false
	}

	foundAll := true
	for _, c := range volCaps {
		if !hasSupport(c) {
			foundAll = false
		}
	}
	return foundAll
}

// getDriverConfigEnforceProxyAccess checks if proxy access is enforced via driver config
func (driver *Driver) getDriverConfigEnforceProxyAccess() bool {
	for k, v := range driver.secrets {
		if strings.ToLower(k) == "enforceproxyaccess" {
			enforce, _ := strconv.ParseBool(v)
			return enforce
		}
	}
	return false
}

// getDriverConfigUser returns user in driver config
func (driver *Driver) getDriverConfigUser() string {
	for k, v := range driver.secrets {
		if strings.ToLower(k) == "user" {
			return v
		}
	}
	return ""
}

// getDriverConfigMountPathWhitelist returns a whitelist of collections that users can mount
func (driver *Driver) getDriverConfigMountPathWhitelist() []string {
	for k, v := range driver.secrets {
		if strings.ToLower(k) == "mountpathwhitelist" {
			whitelist := strings.Split(v, ",")
			for idx := range whitelist {
				whitelist[idx] = strings.TrimSpace(whitelist[idx])
			}

			return whitelist
		}
	}
	return []string{"/"}
}

// isMountPathAllowed checks if given path is allowed to mount
func (driver *Driver) isMountPathAllowed(path string) bool {
	whitelist := driver.getDriverConfigMountPathWhitelist()

	for _, item := range whitelist {
		if checkSubDir(item, path) {
			return true
		}
	}

	return false
}

func checkSubDir(parent string, sub string) bool {
	rel, err := filepath.Rel(parent, sub)
	if err != nil {
		return false
	}

	if !strings.HasPrefix(rel, "..") {
		return true
	}
	return false
}
