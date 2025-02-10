package main

import (
	"fmt"

	"phd.uqtr.ca/lvm-proxy/api"
	"phd.uqtr.ca/lvm-proxy/config"
	"phd.uqtr.ca/lvm-proxy/server"
)

const ONE_GB = 1024 * 1024

func main() {

	// first make sure that pv is there
	err := api.InitializeLVMOnDevice(config.STORAGE_DEVICE)
	if err != nil {
		fmt.Println(err.Error())
	}
	api := api.NewLvmProxyApi(config.STORAGE_DEVICE, config.BRICK_MOUNT_POINT)
	server.StartServer(api)
	// _, err := api.CreateVolumeGroup("test1")
	// if err != nil {
	// 	fmt.Printf("Error creating the volume group %v", err)
	// }
	// lv, err := api.CreateVolume("test-lv", "vg_data", ONE_GB)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// println(lv.AbsolutionPath)
	// err = api.DeleteVolume(lv.Lvo)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// // List volume group
	// vglist := lvm.ListVgNames()
	// // Create a VG object
	// vgo := &lvm.VgObject{}
	// for i := 0; i < len(vglist); i++ {
	// 	vgo.Vgt = lvm.VgOpen(vglist[i], "r")
	// 	vgo.GetSize()
	// 	// if vgo.GetFreeSize() > 0 {
	// 	// 	availableVG = vglist[i]
	// 	// 	vgo.Close()
	// 	// 	break
	// 	// }
	// 	fmt.Printf("VG: %s, Size: %d", vglist[i], vgo.GetSize()/(1000*1000*1000))

	// 	vgo.Close()
	// }
	// lvm.GC()
	// if availableVG == "" {
	// 	fmt.Printf("no VG that has free space found\n")
	// 	return
	// }

	// // Open VG in write mode
	// vgo.Vgt = lvm.VgOpen(availableVG, "w")
	// defer vgo.Close()

	// // Output some data of the VG
	// // fmt.Printf("size: %d GiB\n", uint64(vgo.GetSize())/1024/1024/1024)
	// fmt.Printf("pvlist: %v\n", vgo.ListPVs())
	// // fmt.Printf("Free size: %d KiB\n", uint64(vgo.GetFreeSize())/1024/1024)

	// // Create a LV object
	// l := &lvm.LvObject{}

	// // Create a LV
	// l, err := vgo.CreateLvLinear("go-lvm-example-test-lv", int64(vgo.GetFreeSize())/1024/1024/2)
	// if err != nil {
	// 	fmt.Printf("error: %v")
	// 	return
	// }

	// // Output uuid of LV
	// fmt.Printf("Created\n\tuuid: %s\n\tname: %s\n\tattr: %s\n\torigin: %s\n",
	// 	l.GetUuid(), l.GetName(), l.GetAttr(), l.GetOrigin())

	// // Output uuid of LV
	// l.Remove()
}
