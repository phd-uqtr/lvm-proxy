package api

import "phd.uqtr.ca/lvm-proxy/lvm"

type LvObjectProps struct {
	Lvo                  *lvm.LvObject
	VolumeGroupName      string
	DeviceRelativePath   string
	DeviceAbsolutionPath string
	BrickPath            string
}

type VolumeInfo struct {
	VolumeGroupName    string
	VolumeRelativePath string
	VolumeAbsolutePath string
	BrickPath          string
	Size               uint64
	FreeSize           uint64
	AllocatedSize      uint64
}
