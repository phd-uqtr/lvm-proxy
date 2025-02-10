package api

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"phd.uqtr.ca/lvm-proxy/config"
)

func FormatDevice(device string) error {
	// Check if device exists
	if _, err := os.Stat(device); os.IsNotExist(err) {
		return fmt.Errorf("device %s does not exist", device)
	}
	cmd := exec.Command("mkfs.ext4", "-F", device) // -F forces format without prompt
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Formatting %s as ext4...\n", device)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error formatting disk: %v", err)
	}

	log.Println("Format completed successfully.")
	return nil
}

func MountDevice(device, location string) error {
	// Check if device exists
	if _, err := os.Stat(device); os.IsNotExist(err) {
		return fmt.Errorf("device %s does not exist", device)
	}

	// Create the mount point if it doesn't exist
	if _, err := os.Stat(location); os.IsNotExist(err) {
		log.Printf("Creating mount directory: %s\n", location)
		if err := os.MkdirAll(location, 0755); err != nil {
			return fmt.Errorf("failed to create mount point: %v", err)
		}
	}

	// Mount the device
	cmd := exec.Command("mount", device, location)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Mounting %s to %s...\n", device, location)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error mounting device: %v", err)
	}

	log.Println("Mount successful.")
	return nil
}

func CreateBrick(location string, brick string) (string, error) {
	path := filepath.Join(location, brick)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Creating brick folder: %s\n", path)
		if err := os.MkdirAll(path, 0755); err != nil {
			return "", fmt.Errorf("failed to create brick folder: %v", err)
		}
	}
	return path, nil
}

func InitializeLVMOnDevice(device string) error {
	// TODO: Implement idemtotency
	cmd := exec.Command("pvcreate", device)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error creating pv on device:: %v", err)
	}
	// Create default group
	cmd = exec.Command("vgcreate", config.DEFAULT_GROUP, device)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error creating vg on device:: %v", err)
	}

	return nil
}
