package florist

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
)

// UnzipOne extracts file 'name' from ZIP file 'zipPath' and saves it to 'dstPath'.
func UnzipOne(zipPath string, name string, dstPath string) error {
	log := Log().With("zipPath", zipPath, "name", name, "dstPath", dstPath)
	log.Debug("unzip-one")

	rd, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("UnzipOne: open zip: %s", err)
	}
	defer rd.Close()

	// We calculate the file list also if the log level is not debug. This is somehow
	// a waste; we could use the LogValuer interface as explained in
	// https://pkg.go.dev/log/slog#hdr-Performance_considerations
	// to optimize. But given that we are doing this in the middle of a disk I/O
	// operation, I do not think it is worthwhile.
	files := make([]string, 0, len(rd.File))
	for _, fi := range rd.File {
		files = append(files, fi.Name)
	}
	log.Debug("archive-contents", "files", files)

	fi, err := rd.Open(name)
	if err != nil {
		return fmt.Errorf("UnzipOne: open element: %s", err)
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("UnzipOne: create dst file: %s", err)
	}

	_, err = io.Copy(dst, fi)
	if err != nil {
		return fmt.Errorf("UnzipOne: copy to dst: %s", err)
	}
	log.Debug("unzip-one", "status", "written")

	return nil
}
