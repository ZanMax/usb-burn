//go:build darwin

package formatter

import (
	"fmt"
	"os/exec"
	"strings"

	"usb_burn/internal/device"
)

// diskutilFormatMap maps our format IDs to diskutil eraseDisk format strings.
var diskutilFormatMap = map[string]string{
	"fat32": "FAT32",
	"exfat": "ExFAT",
	"ntfs":  "NTFS",
	"apfs":  "APFS",
	"hfs+":  "HFS+",
}

// AvailableFormats returns filesystem types available on macOS.
func AvailableFormats() []FSType {
	available := []FSType{}
	for _, fs := range AllFormats {
		switch fs.ID {
		case "fat32", "exfat", "apfs", "hfs+":
			// Always available on macOS via diskutil
			available = append(available, fs)
		case "ntfs":
			// Check if ntfs-3g or diskutil can handle it
			// diskutil can create NTFS on some macOS versions
			available = append(available, fs)
		case "ext4":
			// Requires e2fsprogs (brew install e2fsprogs)
			if _, err := exec.LookPath("mkfs.ext4"); err == nil {
				available = append(available, fs)
			} else if _, err := exec.LookPath("/opt/homebrew/sbin/mkfs.ext4"); err == nil {
				available = append(available, fs)
			} else if _, err := exec.LookPath("/usr/local/sbin/mkfs.ext4"); err == nil {
				available = append(available, fs)
			}
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

	var err error
	switch fsType {
	case "ext4":
		err = formatExt4Darwin(drv, label)
	default:
		err = formatDiskutil(drv, fsType, label)
	}

	if err != nil {
		progressCh <- ProgressUpdate{Err: err, Done: true}
		return
	}

	progressCh <- ProgressUpdate{Phase: "Format complete", Done: true}
}

func formatDiskutil(drv *device.Drive, fsType string, label string) error {
	diskutilFmt, ok := diskutilFormatMap[fsType]
	if !ok {
		return fmt.Errorf("unsupported format for diskutil: %s", fsType)
	}

	out, err := exec.Command("diskutil", "eraseDisk", diskutilFmt, label, drv.DeviceNode).CombinedOutput()
	if err != nil {
		return fmt.Errorf("diskutil eraseDisk failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func formatExt4Darwin(drv *device.Drive, label string) error {
	// Try common paths for mkfs.ext4 installed via Homebrew
	mkfsPath := "mkfs.ext4"
	for _, path := range []string{"/opt/homebrew/sbin/mkfs.ext4", "/usr/local/sbin/mkfs.ext4"} {
		if _, err := exec.LookPath(path); err == nil {
			mkfsPath = path
			break
		}
	}

	// Unmount is already done, format the raw device partition
	// First create a single partition using diskutil, then format it
	out, err := exec.Command("diskutil", "eraseDisk", "free", "EMPTY", "MBRFormat", drv.DeviceNode).CombinedOutput()
	if err != nil {
		return fmt.Errorf("disk partitioning failed: %s: %w", strings.TrimSpace(string(out)), err)
	}

	// The partition will be at disk<N>s1
	partition := drv.DeviceNode + "s1"

	out, err = exec.Command(mkfsPath, "-L", label, partition).CombinedOutput()
	if err != nil {
		return fmt.Errorf("mkfs.ext4 failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
