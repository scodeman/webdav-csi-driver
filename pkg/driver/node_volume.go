/*
Original code:
- https://github.com/cyverse/irods-csi-driver/blob/master/pkg/driver/node_volume.go
*/
package driver

// NodeVolume class
type NodeVolume struct {
	ID                        string
	StagingMountPath          string
	MountPath                 string
	DynamicVolumeProvisioning bool
	StageVolume               bool
}

// GetNodeVolume returns NodeVolume with given id
func (driver *Driver) GetNodeVolume(id string) *NodeVolume {
	driver.volumeLock.Lock()
	defer driver.volumeLock.Unlock()

	vol, ok := driver.nodeVolumes[id]
	if !ok {
		return nil
	}
	return vol
}

// PutNodeVolume puts NodeVolume
func (driver *Driver) PutNodeVolume(volume *NodeVolume) {
	driver.volumeLock.Lock()
	defer driver.volumeLock.Unlock()

	driver.nodeVolumes[volume.ID] = volume
}

// PopNodeVolume returns NodeVolume with given id and delete
func (driver *Driver) PopNodeVolume(id string) *NodeVolume {
	driver.volumeLock.Lock()
	defer driver.volumeLock.Unlock()

	vol, ok := driver.nodeVolumes[id]
	if ok {
		delete(driver.nodeVolumes, id)
		return vol
	}
	return nil
}

// CheckNodeVolume returns presence of NodeVolume with given id
func (driver *Driver) CheckNodeVolume(id string) bool {
	driver.volumeLock.Lock()
	defer driver.volumeLock.Unlock()

	_, ok := driver.nodeVolumes[id]
	return ok
}
