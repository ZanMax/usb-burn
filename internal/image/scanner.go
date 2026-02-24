package image

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ImageFile represents a discovered image file.
type ImageFile struct {
	Name string
	Path string
	Size int64
}

// validExtensions are the image file extensions we look for.
var validExtensions = map[string]bool{
	".iso": true,
	".img": true,
	".dmg": true,
	".raw": true,
}

// ScanDirectory scans the given directory for image files.
func ScanDirectory(dir string) ([]ImageFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var images []ImageFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !validExtensions[ext] {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		images = append(images, ImageFile{
			Name: entry.Name(),
			Path: filepath.Join(dir, entry.Name()),
			Size: info.Size(),
		})
	}

	// Sort by name
	sort.Slice(images, func(i, j int) bool {
		return images[i].Name < images[j].Name
	})

	return images, nil
}
