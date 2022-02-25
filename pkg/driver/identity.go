/*
Original code:
- https://github.com/cyverse/irods-csi-driver/blob/master/pkg/driver/identity.go

Following functions or objects are from the code under APL2 License.
- GetPluginInfo
- GetPluginCapabilities
- Probe
Original code: https://github.com/kubernetes-sigs/aws-fsx-csi-driver/blob/master/pkg/driver/identity.go


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

	"github.com/container-storage-interface/spec/lib/go/csi"
)

// GetPluginInfo returns plugin info
func (driver *Driver) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	resp := &csi.GetPluginInfoResponse{
		Name:          driverName,
		VendorVersion: driverVersion,
	}

	return resp, nil
}

// GetPluginCapabilities returns plugin capabilities
func (driver *Driver) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	resp := &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}

	return resp, nil
}

// Probe returns probe response
func (driver *Driver) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}
