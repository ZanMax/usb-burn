//go:build darwin

package device

import (
	"fmt"
	"os/exec"
	"strings"

	"howett.net/plist"
)

// diskutilListOutput represents the plist structure from `diskutil list -plist external`.
type diskutilListOutput struct {
	AllDisks              []string `plist:"AllDisks"`
	AllDisksAndPartitions []struct {
		DeviceIdentifier string `plist:"DeviceIdentifier"`
		Content          string `plist:"Content"`
		Size             int64  `plist:"Size"`
		Partitions       []struct {
			DeviceIdentifier string `plist:"DeviceIdentifier"`
			Content          string `plist:"Content"`
			Size             int64  `plist:"Size"`
			MountPoint       string `plist:"MountPoint"`
		} `plist:"Partitions"`
		MountPoint string `plist:"MountPoint"`
	} `plist:"AllDisksAndPartitions"`
	VolumesFromDisks []string `plist:"VolumesFromDisks"`
	WholeDisks       []string `plist:"WholeDisks"`
}

// diskutilInfoOutput represents the plist from `diskutil info -plist /dev/diskN`.
type diskutilInfoOutput struct {
	DeviceNode          string `plist:"DeviceNode"`
	DeviceIdentifier    string `plist:"DeviceIdentifier"`
	MediaName           string `plist:"MediaName"`
	IORegistryEntryName string `plist:"IORegistryEntryName"`
	Removable           bool   `plist:"Removable"`
	RemovableMedia      bool   `plist:"RemovableMedia"`
	Internal            bool   `plist:"Internal"`
	Size                int64  `plist:"Size"`
	BusProtocol         string `plist:"BusProtocol"`
	VirtualOrPhysical   string `plist:"VirtualOrPhysical"`
	WholeDisk           bool   `plist:"WholeDisk"`
	MountPoint          string `plist:"MountPoint"`
}

// DetectDrives finds all removable external USB drives on macOS.
func DetectDrives() ([]Drive, error) {
	// Get list of external disks
	out, err := exec.Command("diskutil", "list", "-plist", "external").Output()
	if err != nil {
		return nil, fmt.Errorf("diskutil list failed: %w", err)
	}

	var listOutput diskutilListOutput
	_, err = plist.Unmarshal(out, &listOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse diskutil plist: %w", err)
	}

	var drives []Drive
	for _, wholeDisk := range listOutput.WholeDisks {
		deviceNode := "/dev/" + wholeDisk

		// Get detailed info for this disk
		infoOut, err := exec.Command("diskutil", "info", "-plist", deviceNode).Output()
		if err != nil {
			continue // skip drives we can't query
		}

		var info diskutilInfoOutput
		_, err = plist.Unmarshal(infoOut, &info)
		if err != nil {
			continue
		}

		// Filter: only removable, external, physical drives
		if info.Internal {
			continue
		}
		if !info.RemovableMedia && !info.Removable {
			continue
		}
		if info.VirtualOrPhysical == "Virtual" {
			continue
		}

		// Collect mount points from partitions
		var mountPoints []string
		for _, dap := range listOutput.AllDisksAndPartitions {
			if dap.DeviceIdentifier == wholeDisk {
				if dap.MountPoint != "" {
					mountPoints = append(mountPoints, dap.MountPoint)
				}
				for _, p := range dap.Partitions {
					if p.MountPoint != "" {
						mountPoints = append(mountPoints, p.MountPoint)
					}
				}
			}
		}

		name := info.MediaName
		if name == "" {
			name = info.IORegistryEntryName
		}

		drives = append(drives, Drive{
			DeviceNode:  deviceNode,
			RawNode:     "/dev/r" + wholeDisk,
			Name:        name,
			MediaName:   info.MediaName,
			Size:        info.Size,
			Removable:   true,
			MountPoints: mountPoints,
		})
	}

	return drives, nil
}

// UnmountDrive unmounts all volumes on the given drive.
func UnmountDrive(d *Drive) error {
	out, err := exec.Command("diskutil", "unmountDisk", d.DeviceNode).CombinedOutput()
	if err != nil {
		return fmt.Errorf("unmount failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// EjectDrive ejects the given drive.
func EjectDrive(d *Drive) error {
	out, err := exec.Command("diskutil", "eject", d.DeviceNode).CombinedOutput()
	if err != nil {
		return fmt.Errorf("eject failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// IsDiskutilAvailable checks if diskutil is available (always true on macOS).
func IsDiskutilAvailable() bool {
	_, err := exec.LookPath("diskutil")
	return err == nil
}

// GetDiskSize returns the size of a disk using diskutil.
func GetDiskSize(deviceNode string) (int64, error) {
	out, err := exec.Command("diskutil", "info", "-plist", deviceNode).Output()
	if err != nil {
		return 0, err
	}
	var info diskutilInfoOutput
	_, err = plist.Unmarshal(out, &info)
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}
