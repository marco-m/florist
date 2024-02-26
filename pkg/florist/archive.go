package florist

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
)

// UnzipOne extracts file `name` from ZIP file `zipPath` and saves it to `dstPath`.
func UnzipOne(zipPath string, name string, dstPath string) error {
	log := Log.With("zipPath", zipPath, "name", name, "dstPath", dstPath)
	log.Debug("unzip-one")

	rd, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("UnzipOne: open zip: %s", err)
	}
	defer rd.Close()

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
