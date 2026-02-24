package formatter

// FSType represents a supported filesystem type.
type FSType struct {
	ID          string // e.g. "fat32"
	Name        string // e.g. "FAT32"
	Description string // e.g. "Universal, max 4GB file size"
}

// ProgressUpdate carries format progress information for the TUI.
type ProgressUpdate struct {
	Phase string
	Done  bool
	Err   error
}

// AllFormats is the complete list of supported filesystem types.
var AllFormats = []FSType{
	{ID: "fat32", Name: "FAT32", Description: "Universal, max 4GB file size"},
	{ID: "exfat", Name: "exFAT", Description: "Universal, no file size limit"},
	{ID: "ntfs", Name: "NTFS", Description: "Windows native filesystem"},
	{ID: "ext4", Name: "ext4", Description: "Linux native filesystem"},
	{ID: "apfs", Name: "APFS", Description: "macOS native (Apple File System)"},
	{ID: "hfs+", Name: "HFS+", Description: "macOS legacy filesystem"},
}
