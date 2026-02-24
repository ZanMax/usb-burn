//go:build linux

package device

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// lsblkOutput represents the JSON from `lsblk --json --bytes`.
type lsblkOutput struct {
	Blockdevices []lsblkDevice `json:"blockdevices"`
}

type lsblkDevice struct {
	Name       string        `json:"name"`
	Size       int64         `json:"size"`
	Type       string        `json:"type"`
	Removable  string        `json:"rm"`
	MountPoint string        `json:"mountpoint"`
	Model      string        `json:"model"`
	Vendor     string        `json:"vendor"`
	Children   []lsblkDevice `json:"children"`
}

// DetectDrives finds all removable USB drives on Linux.
func DetectDrives() ([]Drive, error) {
	out, err := exec.Command("lsblk", "--json", "--bytes", "-o",
		"NAME,SIZE,TYPE,RM,MOUNTPOINT,MODEL,VENDOR").Output()
	if err != nil {
		return nil, fmt.Errorf("lsblk failed: %w", err)
	}

	var output lsblkOutput
	if err := json.Unmarshal(out, &output); err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %w", err)
	}

	var drives []Drive
	for _, dev := range output.Blockdevices {
		// Filter: only removable whole disks
		if dev.Type != "disk" {
			continue
		}
		if dev.Removable != "1" {
			continue
		}

		var mountPoints []string
		for _, child := range dev.Children {
			if child.MountPoint != "" {
				mountPoints = append(mountPoints, child.MountPoint)
			}
		}
		if dev.MountPoint != "" {
			mountPoints = append(mountPoints, dev.MountPoint)
		}

		name := strings.TrimSpace(dev.Vendor + " " + dev.Model)
		deviceNode := "/dev/" + dev.Name

		drives = append(drives, Drive{
			DeviceNode:  deviceNode,
			RawNode:     deviceNode, // Linux doesn't have raw device nodes
			Name:        name,
			MediaName:   dev.Model,
			Size:        dev.Size,
			Removable:   true,
			MountPoints: mountPoints,
		})
	}

	return drives, nil
}

// UnmountDrive unmounts all partitions on the given drive.
func UnmountDrive(d *Drive) error {
	// Get partitions
	out, err := exec.Command("lsblk", "--json", "-o", "NAME,MOUNTPOINT", d.DeviceNode).Output()
	if err != nil {
		return fmt.Errorf("lsblk failed: %w", err)
	}

	var output lsblkOutput
	if err := json.Unmarshal(out, &output); err != nil {
		return fmt.Errorf("failed to parse lsblk: %w", err)
	}

	for _, dev := range output.Blockdevices {
		for _, child := range dev.Children {
			if child.MountPoint != "" {
				partDev := "/dev/" + child.Name
				umountOut, err := exec.Command("umount", partDev).CombinedOutput()
				if err != nil {
					return fmt.Errorf("unmount %s failed: %s: %w", partDev, strings.TrimSpace(string(umountOut)), err)
				}
			}
		}
	}

	return nil
}

// EjectDrive ejects the given drive on Linux.
func EjectDrive(d *Drive) error {
	out, err := exec.Command("eject", d.DeviceNode).CombinedOutput()
	if err != nil {
		return fmt.Errorf("eject failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
