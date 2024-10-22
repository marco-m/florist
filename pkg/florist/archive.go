package florist

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// UntarOne extracts file 'name' from tar file 'tarPath', expected to be
// compressed with gzip,  and saves it to 'dstPath'.
func UntarOne(tarPath string, name string, dstPath string) error {
	log := Log().With("tarPath", tarPath, "name", name, "dstPath", dstPath)
	log.Debug("untar-one")

	fi, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("UntarOne: %s", err)
	}
	defer fi.Close()

	gzRd, err := gzip.NewReader(fi)
	if err != nil {
		return fmt.Errorf("UntarOne: creating gzip reader for %s: %s", tarPath, err)
	}
	defer gzRd.Close()

	tarRd := tar.NewReader(gzRd)

	var seen []string
	nameFound := false
	for {
		header, err := tarRd.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("UntarOne: reading %s: %s", tarPath, err)
		}
		if !filepath.IsLocal(header.Name) {
			return fmt.Errorf("UntarOne: reading %s: found insecure name %q",
				tarPath, header.Name)
		}
		if header.Name == name {
			nameFound = true
			break
		}
		seen = append(seen, header.Name)
	}
	log.Debug("loop-finished", "names", seen)
	if !nameFound {
		return fmt.Errorf("UntarOne: archive %s does not contain file %s",
			tarPath, name)
	}
	log.Debug("file-found-in-archive")

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("UntarOne: create dst file: %s", err)
	}

	_, err = io.Copy(dst, tarRd)
	if err != nil {
		return fmt.Errorf("UntarOne: copy to dst: %s", err)
	}
	log.Debug("untar-one", "status", "written")

	return nil
}
