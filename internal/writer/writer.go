package writer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"usb_burn/internal/device"
)

// ProgressUpdate carries progress information for the TUI.
type ProgressUpdate struct {
	Phase   string
	Percent float64
	Bytes   int64
	Total   int64
	Speed   float64 // bytes per second
	Done    bool
	Err     error
}

// WriteImage writes the image file to the target drive, sending progress updates on the channel.
func WriteImage(imagePath string, drive *device.Drive, progressCh chan<- ProgressUpdate) {
	defer close(progressCh)

	// Open image file
	progressCh <- ProgressUpdate{Phase: "Opening image..."}
	imgFile, err := os.Open(imagePath)
	if err != nil {
		progressCh <- ProgressUpdate{Err: fmt.Errorf("failed to open image: %w", err), Done: true}
		return
	}
	defer imgFile.Close()

	imgInfo, err := imgFile.Stat()
	if err != nil {
		progressCh <- ProgressUpdate{Err: fmt.Errorf("failed to stat image: %w", err), Done: true}
		return
	}
	totalBytes := imgInfo.Size()

	// Check image fits on drive
	if totalBytes > drive.Size {
		progressCh <- ProgressUpdate{
			Err:  fmt.Errorf("image (%s) is larger than drive (%s)", device.FormatBytes(totalBytes), device.FormatBytes(drive.Size)),
			Done: true,
		}
		return
	}

	// Unmount drive
	progressCh <- ProgressUpdate{Phase: "Unmounting drive..."}
	if err := device.UnmountDrive(drive); err != nil {
		progressCh <- ProgressUpdate{Err: fmt.Errorf("failed to unmount: %w", err), Done: true}
		return
	}

	// Open raw device for writing (no O_SYNC — buffered writes are much faster;
	// we flush once at the end, same as dd)
	progressCh <- ProgressUpdate{Phase: "Opening device..."}
	devFile, err := os.OpenFile(drive.RawNode, os.O_WRONLY, 0)
	if err != nil {
		progressCh <- ProgressUpdate{Err: fmt.Errorf("failed to open device %s: %w", drive.RawNode, err), Done: true}
		return
	}
	defer devFile.Close()

	// Write with progress tracking
	progressCh <- ProgressUpdate{Phase: "Writing image...", Total: totalBytes}

	cr := NewCountingReader(imgFile)
	buf := make([]byte, 4*1024*1024) // 4MB buffer

	startTime := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Progress reporting goroutine
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(startTime).Seconds()
				bytesRead := cr.BytesRead
				var speed float64
				if elapsed > 0 {
					speed = float64(bytesRead) / elapsed
				}
				var pct float64
				if totalBytes > 0 {
					pct = float64(bytesRead) / float64(totalBytes)
				}
				progressCh <- ProgressUpdate{
					Phase:   "Writing image...",
					Percent: pct,
					Bytes:   bytesRead,
					Total:   totalBytes,
					Speed:   speed,
				}
			case <-done:
				return
			}
		}
	}()

	_, err = io.CopyBuffer(devFile, cr, buf)
	close(done)

	if err != nil {
		progressCh <- ProgressUpdate{Err: fmt.Errorf("write failed: %w", err), Done: true}
		return
	}

	// Sync — flush any buffered writes to the device.
	// On macOS, Sync() on raw disk devices (/dev/rdiskN) fails with ENOTTY
	// because F_FULLFSYNC/fsync aren't supported on raw devices. This is safe
	// to ignore: closing the fd will flush remaining kernel buffers.
	progressCh <- ProgressUpdate{Phase: "Syncing...", Percent: 1.0, Bytes: totalBytes, Total: totalBytes}
	if err := devFile.Sync(); err != nil {
		if !errors.Is(err, syscall.ENOTTY) {
			progressCh <- ProgressUpdate{Err: fmt.Errorf("sync failed: %w", err), Done: true}
			return
		}
	}

	// Eject
	progressCh <- ProgressUpdate{Phase: "Ejecting..."}
	devFile.Close()
	_ = device.EjectDrive(drive)

	elapsed := time.Since(startTime).Seconds()
	var speed float64
	if elapsed > 0 {
		speed = float64(totalBytes) / elapsed
	}

	progressCh <- ProgressUpdate{
		Phase:   "Complete",
		Percent: 1.0,
		Bytes:   totalBytes,
		Total:   totalBytes,
		Speed:   speed,
		Done:    true,
	}
}
