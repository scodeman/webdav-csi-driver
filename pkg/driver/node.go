/*
Original code:
- https://github.com/cyverse/irods-csi-driver/blob/master/pkg/driver/node.go

Following functions or objects are from the code under APL2 License.
- NodeStageVolume
- NodePublishVolume
- NodeUnpublishVolume
- NodeUnstageVolume
- NodeGetCapabilities
- NodeGetInfo
Original code:
- https://github.com/kubernetes-sigs/aws-efs-csi-driver/blob/master/pkg/driver/node.go
- https://github.com/kubernetes-sigs/aws-fsx-csi-driver/blob/master/pkg/driver/node.go


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
	"fmt"
	"os"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

var (
	nodeCaps = []csi.NodeServiceCapability_RPC_Type{csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME}
)

// NodeStageVolume handles persistent volume stage event in node service
func (driver *Driver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	klog.V(4).Infof("NodeStageVolume: volumeId (%#v)", volID)

	if !driver.isDynamicVolumeProvisioningMode(req.VolumeContext) {
		// static volume provisioning
		nodeVolume := &NodeVolume{
			ID:                        volID,
			StagingMountPath:          "",
			MountPath:                 "",
			DynamicVolumeProvisioning: false,
			StageVolume:               true,
		}
		driver.PutNodeVolume(nodeVolume)

		return &csi.NodeStageVolumeResponse{}, nil
	}

	// only for dynamic volume provisioning mode
	targetPath := req.GetStagingTargetPath()
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Staging target path not provided")
	}

	volCap := req.GetVolumeCapability()
	if volCap == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability not provided")
	}

	if !driver.isValidVolumeCapabilities([]*csi.VolumeCapability{volCap}) {
		return nil, status.Error(codes.InvalidArgument, "Volume capability not supported")
	}

	mountOptions := []string{}
	if m := volCap.GetMount(); m != nil {
		hasOption := func(options []string, opt string) bool {
			for _, o := range options {
				if o == opt {
					return true
				}
			}
			return false
		}
		for _, f := range m.MountFlags {
			if !hasOption(mountOptions, f) {
				mountOptions = append(mountOptions, f)
			}
		}
	}

	pathExist, pathExistErr := PathExists(targetPath)
	if pathExistErr != nil {
		return nil, status.Error(codes.Internal, pathExistErr.Error())
	}

	if !pathExist {
		klog.V(5).Infof("NodeStageVolume: creating dir %s", targetPath)
		if err := MakeDir(targetPath); err != nil {
			return nil, status.Errorf(codes.Internal, "Could not create dir %q: %v", targetPath, err)
		}
	}

	notMountPoint, err := driver.mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !notMountPoint {
		return nil, status.Errorf(codes.Internal, "Staging target path %s is already mounted", targetPath)
	}

	volContext := req.GetVolumeContext()
	volSecrets := req.GetSecrets()

	secrets := make(map[string]string)
	for k, v := range driver.secrets {
		secrets[k] = v
	}

	for k, v := range volSecrets {
		secrets[k] = v
	}

	client := ExtractClientType(volContext, secrets, WebdavType)

	switch client {
	case WebdavType:
		klog.V(5).Infof("NodeStageVolume: mounting %s", client)
		if err := driver.mountWebdav(volContext, secrets, mountOptions, targetPath); err != nil {
			os.Remove(targetPath)
			return nil, err
		}
	default:
		return nil, status.Errorf(codes.Internal, "unknown driver type - %v", client)
	}

	klog.V(5).Infof("NodeStageVolume: %s was mounted", targetPath)

	nodeVolume := &NodeVolume{
		ID:                        volID,
		StagingMountPath:          targetPath,
		MountPath:                 "",
		DynamicVolumeProvisioning: true,
		StageVolume:               true,
	}
	driver.PutNodeVolume(nodeVolume)

	return &csi.NodeStageVolumeResponse{}, nil
}

// NodePublishVolume handles persistent volume publish event in node service
func (driver *Driver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	klog.V(4).Infof("NodePublishVolume: volumeId (%#v)", volID)

	targetPath := req.GetTargetPath()
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path not provided")
	}

	volCap := req.GetVolumeCapability()
	if volCap == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability not provided")
	}

	if !driver.isValidVolumeCapabilities([]*csi.VolumeCapability{volCap}) {
		return nil, status.Error(codes.InvalidArgument, "Volume capability not supported")
	}

	mountOptions := []string{}
	if req.GetReadonly() {
		mountOptions = append(mountOptions, "ro")
	}

	if m := volCap.GetMount(); m != nil {
		hasOption := func(options []string, opt string) bool {
			for _, o := range options {
				if o == opt {
					return true
				}
			}
			return false
		}
		for _, f := range m.MountFlags {
			if !hasOption(mountOptions, f) {
				mountOptions = append(mountOptions, f)
			}
		}
	}

	pathExist, pathExistErr := PathExists(targetPath)
	if pathExistErr != nil {
		return nil, status.Error(codes.Internal, pathExistErr.Error())
	}

	if !pathExist {
		klog.V(5).Infof("NodePublishVolume: creating dir %s", targetPath)
		if err := MakeDir(targetPath); err != nil {
			return nil, status.Errorf(codes.Internal, "Could not create dir %q: %v", targetPath, err)
		}
	}

	notMountPoint, err := driver.mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !notMountPoint {
		return nil, status.Errorf(codes.Internal, "Staging target path %s is already mounted", targetPath)
	}

	if driver.isDynamicVolumeProvisioningMode(req.VolumeContext) {
		// dynamic volume provisioning
		// bind mount
		stagingTargetPath := req.GetStagingTargetPath()
		if len(stagingTargetPath) == 0 {
			return nil, status.Error(codes.InvalidArgument, "Staging target path not provided")
		}

		klog.V(5).Infof("NodePublishVolume: mounting %s", "bind")
		if err := driver.mountBind(stagingTargetPath, mountOptions, targetPath); err != nil {
			os.Remove(targetPath)
			return nil, err
		}

		// update node volume info
		nodeVolume := driver.PopNodeVolume(volID)
		if nodeVolume == nil {
			return nil, status.Errorf(codes.InvalidArgument, "Unable to find node volume %s", volID)
		}

		nodeVolume.MountPath = targetPath
		driver.PutNodeVolume(nodeVolume)
	} else {
		// static volume provisioning
		// mount volume
		volContext := req.GetVolumeContext()
		volSecrets := req.GetSecrets()

		secrets := make(map[string]string)
		for k, v := range driver.secrets {
			secrets[k] = v
		}

		for k, v := range volSecrets {
			secrets[k] = v
		}

		client := ExtractClientType(volContext, secrets, WebdavType)

		switch client {
		case WebdavType:
			klog.V(5).Infof("NodePublishVolume: mounting %s", client)
			if err := driver.mountWebdav(volContext, secrets, mountOptions, targetPath); err != nil {
				os.Remove(targetPath)
				return nil, err
			}
		default:
			return nil, status.Errorf(codes.Internal, "unknown driver type - %v", client)
		}

		// update node volume info if exists
		nodeVolume := driver.PopNodeVolume(volID)
		if nodeVolume == nil {
			nodeVolume = &NodeVolume{
				ID:                        volID,
				StagingMountPath:          "",
				MountPath:                 targetPath,
				DynamicVolumeProvisioning: false,
				StageVolume:               false,
			}
		} else {
			nodeVolume.MountPath = targetPath
		}

		driver.PutNodeVolume(nodeVolume)
	}

	klog.V(5).Infof("NodePublishVolume: %s was mounted", targetPath)

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume handles persistent volume unpublish event in node service
func (driver *Driver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	klog.V(4).Infof("NodeUnpublishVolume: volumeId (%#v)", volID)

	nodeVolume := driver.GetNodeVolume(volID)
	if nodeVolume == nil {
		klog.V(5).Infof("Unable to find node volume %s", volID)
	} else {
		if !nodeVolume.StageVolume {
			// delete here
			driver.PopNodeVolume(volID)
		}
	}

	targetPath := req.GetTargetPath()
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path not provided")
	}

	// Check if target directory is a mount point. GetDeviceNameFromMount
	// given a mnt point, finds the device from /proc/mounts
	// returns the device name, reference count, and error code
	_, refCount, err := driver.mounter.GetDeviceName(targetPath)
	if err != nil {
		msg := fmt.Sprintf("failed to check if volume is mounted: %v", err)
		return nil, status.Error(codes.Internal, msg)
	}

	// From the spec: If the volume corresponding to the volume_id
	// is not staged to the staging_target_path, the Plugin MUST
	// reply 0 OK.
	if refCount == 0 {
		klog.V(5).Infof("NodeUnpublishVolume: %s target not mounted", targetPath)
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	klog.V(5).Infof("NodeUnpublishVolume: unmounting %s", targetPath)
	// unmount
	err = driver.mounter.UnmountForcefully(targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not unmount %q: %v", targetPath, err)
	}
	klog.V(5).Infof("NodeUnpublishVolume: %s unmounted", targetPath)

	err = os.Remove(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeUnstageVolume handles persistent volume unstage event in node service
func (driver *Driver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	klog.V(4).Infof("NodeUnstageVolume: volumeId (%#v)", volID)

	nodeVolume := driver.GetNodeVolume(volID)
	if nodeVolume == nil {
		klog.V(5).Infof("Unable to find node volume %s", volID)
	} else {
		// delete here
		driver.PopNodeVolume(volID)

		if !nodeVolume.DynamicVolumeProvisioning {
			// nothing to do for StaticCVolumeProvisioning
			return &csi.NodeUnstageVolumeResponse{}, nil
		}
	}

	targetPath := req.GetStagingTargetPath()
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Staging target path not provided")
	}

	// Check if target directory is a mount point. GetDeviceNameFromMount
	// given a mnt point, finds the device from /proc/mounts
	// returns the device name, reference count, and error code
	_, refCount, err := driver.mounter.GetDeviceName(targetPath)
	if err != nil {
		msg := fmt.Sprintf("failed to check if volume is mounted: %v", err)
		return nil, status.Error(codes.Internal, msg)
	}

	// From the spec: If the volume corresponding to the volume_id
	// is not staged to the staging_target_path, the Plugin MUST
	// reply 0 OK.
	if refCount == 0 {
		klog.V(5).Infof("NodeUnstageVolume: %s target not mounted", targetPath)
		return &csi.NodeUnstageVolumeResponse{}, nil
	}

	klog.V(5).Infof("NodeUnstageVolume: unmounting %s", targetPath)
	err = driver.mounter.UnmountForcefully(targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not unmount %q: %v", targetPath, err)
	}
	klog.V(5).Infof("NodeUnstageVolume: %s unmounted", targetPath)

	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodeGetVolumeStats returns volume stats
func (driver *Driver) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeExpandVolume expands volume
func (driver *Driver) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeGetCapabilities returns capabilities
func (driver *Driver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	klog.V(4).Infof("NodeGetCapabilities: called with args %+v", req)
	var caps []*csi.NodeServiceCapability
	for _, cap := range nodeCaps {
		c := &csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: cap,
				},
			},
		}
		caps = append(caps, c)
	}
	return &csi.NodeGetCapabilitiesResponse{Capabilities: caps}, nil
}

// NodeGetInfo returns node info
func (driver *Driver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	klog.V(4).Infof("NodeGetInfo: called with args %+v", req)

	return &csi.NodeGetInfoResponse{
		NodeId: driver.config.NodeID,
	}, nil
}

func (driver *Driver) mountBind(sourcePath string, mntOptions []string, targetPath string) error {
	fsType := ""
	mountOptions := []string{}
	mountSensitiveOptions := []string{}
	stdinArgs := []string{}

	mountOptions = append(mountOptions, mntOptions...)
	mountOptions = append(mountOptions, "bind")

	klog.V(5).Infof("Mounting %s at %s with options %v", sourcePath, targetPath, mountOptions)
	if err := driver.mounter.MountSensitive2(sourcePath, sourcePath, targetPath, fsType, mountOptions, mountSensitiveOptions, stdinArgs); err != nil {
		return status.Errorf(codes.Internal, "Could not mount %q (%q) at %q: %v", sourcePath, fsType, targetPath, err)
	}

	return nil
}

func (driver *Driver) mountWebdav(volContext map[string]string, volSecrets map[string]string, mntOptions []string, targetPath string) error {
	conn, err := ExtractWebDAVConnectionInfo(volContext, volSecrets)
	if err != nil {
		return err
	}

	fsType := "davfs"
	source := conn.URL

	mountOptions := []string{}
	mountSensitiveOptions := []string{}
	stdinArgs := []string{}

	mountOptions = append(mountOptions, mntOptions...)

	// if user == anonymous, password is empty, and doesn't need to pass user/password as arguments
	if len(conn.User) > 0 && conn.User != "anonymous" && len(conn.Password) > 0 {
		mountSensitiveOptions = append(mountSensitiveOptions, fmt.Sprintf("username=%s", conn.User))
		stdinArgs = append(stdinArgs, conn.Password)
	}

	klog.V(5).Infof("Mounting %s (%s) at %s with options %v", source, fsType, targetPath, mountOptions)
	if err := driver.mounter.MountSensitive2(source, source, targetPath, fsType, mountOptions, mountSensitiveOptions, stdinArgs); err != nil {
		return status.Errorf(codes.Internal, "Could not mount %q (%q) at %q: %v", source, fsType, targetPath, err)
	}

	return nil
}

func (driver *Driver) isDynamicVolumeProvisioningMode(volContext map[string]string) bool {
	for k, v := range volContext {
		if strings.ToLower(k) == "provisioning_mode" {
			if strings.ToLower(v) == "dynamic" {
				return true
			}
		}
	}

	return false
}
