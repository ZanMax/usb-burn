package device

import "fmt"

// Drive represents a detected USB drive.
type Drive struct {
	DeviceNode  string   // e.g. /dev/disk2
	RawNode     string   // e.g. /dev/rdisk2 (macOS) or same as DeviceNode (Linux)
	Name        string   // vendor/model name
	MediaName   string   // media-specific name
	Size        int64    // size in bytes
	Removable   bool
	MountPoints []string
}

// DisplayName returns a human-readable name for the drive.
func (d Drive) DisplayName() string {
	name := d.Name
	if name == "" {
		name = d.MediaName
	}
	if name == "" {
		name = "Unknown Drive"
	}
	return name
}

// SizeHuman returns the drive size in a human-readable format.
func (d Drive) SizeHuman() string {
	return FormatBytes(d.Size)
}

// String implements fmt.Stringer.
func (d Drive) String() string {
	return fmt.Sprintf("%s (%s) [%s]", d.DisplayName(), d.SizeHuman(), d.DeviceNode)
}

// FormatBytes converts bytes to a human-readable string.
func FormatBytes(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
