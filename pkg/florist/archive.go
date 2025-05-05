package florist

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/marco-m/florist/pkg/sets"
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
// compressed with gzip, and saves it to 'dstPath'.
//
// Examples:
//
//   - Flat archive, extract file "foo":              UntarOne(tarPath, "foo", dst)
//   - Hierarchical archive, extract file "bar/foo":  UntarOne(tarPath, "bar/foo", dst)
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
			return fmt.Errorf("UntarOne: reading %s: found insecure path %q",
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

// UntarSome extracts the files in tar file 'tarPath' that match the list in 'some'. File
// 'tarPath' is expected to be compressed with gzip. UntarSome saves the matching files to
// directory 'dstDir', which must exist beforehand, and applies to them 'perm', 'owner'
// and 'group'. UntarSome disregards the file permissions and ownrship that are present in
// the tarfile. If any file in 'some' is not found, UntarSome returns an error. In case
// the archive is hierarchical, UntarSome flattens any directory. This behavior might
// change. See also [UntarAll].
func UntarSome(tarPath, dstDir string, some []string, perm os.FileMode, owner, group string) error {
	fn := "UntarSome"
	if some == nil {
		fn = "UntarAll"
	}
	log := Log().With("fn", fn)
	errorf := makeErrorf(fn)

	log.Debug(fn, "phase", "starting", "tarPath", tarPath, "dstDir", dstDir, "some", some)

	fi, err := os.Open(tarPath)
	if err != nil {
		return errorf("%w", err)
	}
	defer fi.Close()

	gzRd, err := gzip.NewReader(fi)
	if err != nil {
		return errorf("creating gzip reader for %s: %s", tarPath, err)
	}
	defer gzRd.Close()

	tarRd := tar.NewReader(gzRd)
	wanted := sets.From(some...)
	found := sets.New[string](wanted.Size())

	for {
		header, err := tarRd.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return errorf("reading %s: %s", tarPath, err)
		}
		if !filepath.IsLocal(header.Name) {
			return errorf("reading %s: found insecure path %q", tarPath, header.Name)
		}
		if header.Typeflag != tar.TypeReg {
			// Skip any non-regular file.
			continue
		}
		if wanted.Size() > 0 {
			// Depending on how the tarfile has been created, the same relative path "foo"
			// could be encoded in two ways: "header.Name = foo" or "header.Name = ./foo".
			// We need to normalize.
			canonicalName := filepath.Clean(header.Name)
			if !wanted.Contains(canonicalName) {
				log.Debug(fn, "skip", header.Name, "canonical", canonicalName)
				continue
			} else {
				found.Add(canonicalName)
			}
		}

		dstPath := filepath.Join(dstDir, header.Name)
		log.Debug(fn, "phase", "copying", "dst", dstPath)
		dst, err := os.Create(dstPath)
		if err != nil {
			return errorf("creating dst file: %s", err)
		}
		if err := ChOwnMod(dstPath, perm, owner, group); err != nil {
			return errorf("changing owner/mode for file: %s", err)
		}

		_, err = io.Copy(dst, tarRd)
		if err != nil {
			return errorf("copy to dst: %s", err)
		}
		log.Info(fn, "phase", "written", "dst", dstPath)
	}
	if wanted.Size() > 0 {
		notFound := wanted.Difference(found)
		if notFound.Size() > 0 {
			return errorf("some files not found: %s", notFound)
		}
	}

	return nil
}

// UntarAll extracts all files from tar file 'tarPath'. File 'tarPath' is expected to be
// compressed with gzip. UntarSome saves the matching files to directory 'dstDir', which
// must exist beforehand, and applies to them 'perm', 'owner' and 'group'. UntarAll
// disregards the file permissions and ownrship that are present in the tarfile. In case
// the archive is hierarchical, UntarAll flattens any directory. This behavior might
// change. See also [UntarSome].
func UntarAll(tarPath, dstDir string, perm os.FileMode, owner, group string) error {
	return UntarSome(tarPath, dstDir, nil, perm, owner, group)
}
