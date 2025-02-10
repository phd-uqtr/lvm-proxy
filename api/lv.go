package api

import "phd.uqtr.ca/lvm-proxy/lvm"

type LvObjectProps struct {
	Lvo                  *lvm.LvObject
	VolumeGroupName      string
	DeviceRelativePath   string
	DeviceAbsolutionPath string
	BrickPath            string
}
