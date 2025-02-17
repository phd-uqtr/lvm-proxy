package api

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/sys/unix"
	"phd.uqtr.ca/lvm-proxy/lvm"
)

const ONE_GB = 1024 * 1024

// TODO: We might extend this to vg with multiple base storages
type LvmProxyApi struct {
	BaseStorageDevice string // path to base storage used for creating logical volumes on to of it
	MountRoot         string
}

func (lvmApi *LvmProxyApi) GetVolumeMountPath(vgName, lvName string) string {
	return path.Join(lvmApi.MountRoot, vgName, lvName)
}

func (lvmApi *LvmProxyApi) GetVolumeGroupInfo(vg string) VolumeGroupInfo {
	vgo := &lvm.VgObject{}
	vgo.Vgt = lvm.VgOpen(vg, "r")
	defer vgo.Close()
	vgInfo := VolumeGroupInfo{
		Name:     vg,
		Size:     uint64(vgo.GetSize()),
		FreeSize: uint64(vgo.GetFreeSize()),
	}
	return vgInfo
}

func (lvmApi *LvmProxyApi) GetVolumeGroupNames() []string {
	return lvm.ListVgNames()
}

func (lvmApi *LvmProxyApi) CreateVolumeGroup(vgName string) (*lvm.VgObject, error) {
	// TODO: Check if vg exists
	vgo := lvm.VgCreate(vgName)
	err := vgo.Extend(lvmApi.BaseStorageDevice)
	return vgo, err
}

func (lvmApi *LvmProxyApi) GetVolumes(vgName string) ([]LvObjectProps, error) {
	// Open volume
	vgo := &lvm.VgObject{}
	vgo.Vgt = lvm.VgOpen(vgName, "r")
	defer vgo.Close()
	if vgo.Vgt == nil {
		return nil, fmt.Errorf("no volume of name: %s", vgName)
	}
	var lvs = make([]LvObjectProps, 0)
	lvNames := vgo.ListLVs()
	for _, lvName := range lvNames {
		lv, err := vgo.LvFromName(lvName)
		if err != nil {
			// fmt.Println(err)
			continue
		}
		path := fmt.Sprintf("/dev/%s/%s", vgName, lvName)
		absPath, err := os.Readlink(path)
		if err != nil {
			absPath = ""
		}
		brickPath := filepath.Join(lvmApi.MountRoot, vgName, lvName, "brick")
		lvProps := LvObjectProps{
			Lvo:                  lv,
			DeviceRelativePath:   path,
			DeviceAbsolutionPath: absPath,
			VolumeGroupName:      vgName,
			BrickPath:            brickPath,
		}
		lvs = append(lvs, lvProps)
	}
	return lvs, nil

}

func (lvmApi *LvmProxyApi) CreateVolume(volName string, vgName string, volSize int64) (*LvObjectProps, error) {

	// Create a VG object
	vgo := &lvm.VgObject{}
	vgo.Vgt = lvm.VgOpen(vgName, "w")
	defer vgo.Close()

	// Create a LV object
	lv := &lvm.LvObject{}

	// Create LV
	lv, err := vgo.CreateLvLinear(volName, volSize)
	if err != nil {
		// fmt.Printf("Error creating the volume\n")
		return nil, fmt.Errorf("error creating the volume: %v", err)
	}

	path := filepath.Join("/dev", vgName, volName)
	absPath, err := os.Readlink(path)
	if err != nil {
		absPath = ""
	}
	// - after creating the volume we should format /dev/mapper/vg-vol as ext4
	// - next we mount the formatted device under a folder
	// - next we create the `brick` folder in the mounted
	// - we should return the brick path to the client

	err = FormatDevice(absPath)
	if err != nil {
		// TODO: Remove volume
		return nil, err
	}

	// Mount location
	mountPath := filepath.Join(lvmApi.MountRoot, vgName, volName)
	err = MountDevice(absPath, mountPath)
	if err != nil {
		// TODO: Remove volume and mount point
		return nil, err
	}
	brickPath, err := CreateBrick(mountPath, "brick")
	if err != nil {
		// TODO: Remove volume and mount point
		return nil, err
	}

	props := &LvObjectProps{
		Lvo:                  lv,
		DeviceRelativePath:   path,
		DeviceAbsolutionPath: absPath,
		VolumeGroupName:      vgName,
		BrickPath:            brickPath,
	}

	return props, nil
}

func (lvmApi *LvmProxyApi) GetVolumeInfo(vgName, lvName string) *VolumeInfo {
	// TODO: Check if volume and volume groups exist
	lvPath := path.Join("/dev/mapper", fmt.Sprintf("%s-%s", vgName, lvName))
	lvAbsPath, err := os.Readlink(lvPath)
	if err != nil {
		lvAbsPath = ""
	}
	mountPath := filepath.Join(lvmApi.MountRoot, vgName, lvName)
	brickPath := path.Join(mountPath, "brick")
	var lvStats unix.Statfs_t
	err = unix.Statfs(lvPath, &lvStats)
	if err != nil {
		return nil
	}
	total := lvStats.Blocks * uint64(lvStats.Bsize)
	available := lvStats.Bavail * uint64(lvStats.Bsize)
	used := total - (lvStats.Bfree * uint64(lvStats.Bsize))
	return &VolumeInfo{
		VolumeGroupName:    vgName,
		VolumeRelativePath: lvPath,
		VolumeAbsolutePath: lvAbsPath,
		BrickPath:          brickPath,
		Size:               total,
		FreeSize:           available,
		AllocatedSize:      used,
	}
}

func (lvmApi *LvmProxyApi) DeleteVolume(vgName, lvName string) error {
	// Check if vg exists
	exists, vgo := lvm.VgExists(vgName, "w")
	if !exists {
		return fmt.Errorf("vg does not exist")
	}
	defer vgo.Close()
	// Check if lv exists
	lvo, err := vgo.LvFromName(lvName)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Check if lv is mounted
	lvPath := path.Join("/dev/mapper", fmt.Sprintf("%s-%s", vgName, lvName))
	isMounted := IsVolumeMounted(lvPath)
	if !isMounted {
		return lvo.Remove()
	}
	// unmout first and remove folder\
	mountPath := lvmApi.GetVolumeMountPath(vgName, lvName)
	err = Unmount(mountPath)
	if err != nil {
		return err
	}
	err = os.RemoveAll(mountPath)
	if err != nil {
		return fmt.Errorf("failed to remove %s: %v", mountPath, err)
	}
	// Delete volume
	return lvo.Remove()
}

func NewLvmProxyApi(baseDevice string, mountRoot string) *LvmProxyApi {

	return &LvmProxyApi{
		BaseStorageDevice: baseDevice,
		MountRoot:         mountRoot,
	}
}
