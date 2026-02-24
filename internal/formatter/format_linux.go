//go:build linux

package formatter

import (
	"fmt"
	"os/exec"
	"strings"

	"usb_burn/internal/device"
)

// AvailableFormats returns filesystem types available on Linux.
func AvailableFormats() []FSType {
	available := []FSType{}
	for _, fs := range AllFormats {
		switch fs.ID {
		case "fat32":
			if _, err := exec.LookPath("mkfs.vfat"); err == nil {
				available = append(available, fs)
			}
		case "exfat":
			if _, err := exec.LookPath("mkfs.exfat"); err == nil {
				available = append(available, fs)
			}
		case "ntfs":
			if _, err := exec.LookPath("mkfs.ntfs"); err == nil {
				available = append(available, fs)
			}
		case "ext4":
			if _, err := exec.LookPath("mkfs.ext4"); err == nil {
				available = append(available, fs)
			}
		case "hfs+":
			if _, err := exec.LookPath("mkfs.hfsplus"); err == nil {
				available = append(available, fs)
			}
		case "apfs":
			// APFS is not supported on Linux
		}
	}
	return available
}

// FormatDrive formats the given drive with the specified filesystem.
func FormatDrive(drv *device.Drive, fsType string, label string, progressCh chan<- ProgressUpdate) {
	defer close(progressCh)

	if label == "" {
		label = "USB_BURN"
	}

	// Unmount
	progressCh <- ProgressUpdate{Phase: "Unmounting drive..."}
	if err := device.UnmountDrive(drv); err != nil {
		progressCh <- ProgressUpdate{Err: fmt.Errorf("unmount failed: %w", err), Done: true}
		return
	}

	progressCh <- ProgressUpdate{Phase: fmt.Sprintf("Formatting as %s...", strings.ToUpper(fsType))}

	var cmd *exec.Cmd
	partition := drv.DeviceNode + "1" // e.g. /dev/sdb1

	// First, create a new partition table
	wipeCmds := fmt.Sprintf("echo 'o\nn\np\n1\n\n\nw' | fdisk %s", drv.DeviceNode)
	wipeOut, err := exec.Command("bash", "-c", wipeCmds).CombinedOutput()
	if err != nil {
		progressCh <- ProgressUpdate{
			Err:  fmt.Errorf("partition table creation failed: %s: %w", strings.TrimSpace(string(wipeOut)), err),
			Done: true,
		}
		return
	}

	switch fsType {
	case "fat32":
		cmd = exec.Command("mkfs.vfat", "-F", "32", "-n", label, partition)
	case "exfat":
		cmd = exec.Command("mkfs.exfat", "-n", label, partition)
	case "ntfs":
		cmd = exec.Command("mkfs.ntfs", "--fast", "-L", label, partition)
	case "ext4":
		cmd = exec.Command("mkfs.ext4", "-L", label, partition)
	case "hfs+":
		cmd = exec.Command("mkfs.hfsplus", "-v", label, partition)
	default:
		progressCh <- ProgressUpdate{Err: fmt.Errorf("unsupported format: %s", fsType), Done: true}
		return
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		progressCh <- ProgressUpdate{
			Err:  fmt.Errorf("format failed: %s: %w", strings.TrimSpace(string(out)), err),
			Done: true,
		}
		return
	}

	progressCh <- ProgressUpdate{Phase: "Format complete", Done: true}
}
