package florist

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path"
	"strconv"
	"text/template"
)

// Mkdir creates the directory path with associated owner and permissions.
// If the directory already exists it does nothing.
func Mkdir(path string, owner *user.User, perm fs.FileMode) error {
	log := Log.Named("Mkdir").With("path", path)

	_, err := os.Stat(path)
	if err == nil {
		log.Debug("directory exists")
		return nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		// Some other error
		return fmt.Errorf("florist.Mkdir: %s", err)
	}
	// Directory doesn't exist, let's create it
	log.Debug("creating directory", "dir", path, "perm", perm)
	if err := os.Mkdir(path, perm); err != nil {
		return fmt.Errorf("florist.Mkdir: %s", err)
	}

	if owner != nil {
		uid, _ := strconv.Atoi(owner.Uid)
		gid, _ := strconv.Atoi(owner.Gid)
		log.Debug("chowning directory", "owner", owner.Username)
		if err := os.Chown(path, uid, gid); err != nil {
			return fmt.Errorf("florist.Mkdir: %s", err)
		}
	}

	return nil
}

// CopyFile copies file srcPath to dstPath, with mode and owner. The source and
// destination files reside in the "real" filesystem.
// Notes:
// - If dstPath exists, it will be overwritten.
// - Setting an owner different that the current user requires elevated privileges.
func CopyFile(
	srcPath string, dstPath string,
	mode os.FileMode, owner *user.User,
) error {
	return copyfile(nil, srcPath, dstPath, mode, owner)
}

// CopyFileFromFs copies file srcPath to dstPath, with mode and owner. The source
// file resides in the srcFs filesystem (for example, via go:embed), while the
// destination file resides in the "real" filesystem.
// Notes:
// - If dstPath exists, it will be overwritten.
// - Setting an owner different that the current user requires elevated privileges.
func CopyFileFromFs(
	srcFs fs.FS, srcPath string, dstPath string,
	mode os.FileMode, owner *user.User,
) error {
	return copyfile(srcFs, srcPath, dstPath, mode, owner)
}

// CopyTemplateFromFs copies file srcPath to dstPath, with mode and owner, performing
// Go text template processing based on the fields in tmplData. The source file resides
// in the srcFs filesystem (for example, via go:embed), while the destination file
// resides in the "real" filesystem.
// Notes:
//   - If dstPath exists, it will be overwritten.
//   - Setting an owner different that the current user requires elevated privileges.
//   - Template delimiters delimL, delimR allow to easily support templates that contain
//     themselves the default delimiters {{, }}. Passing the empty delimiters stand for
//     the default.
func CopyTemplateFromFs(
	srcFs fs.FS, srcPath string, dstPath string,
	mode os.FileMode, owner *user.User, tmplData any, delimL, delimR string,
) error {
	return copytemplate(srcFs, srcPath, dstPath, mode, owner, tmplData, delimL, delimR)
}

// Does not read the whole file in memory.
func copyfile(
	srcFs fs.FS, srcPath string, dstPath string,
	mode os.FileMode, owner *user.User,
) error {
	var src fs.File
	var err error
	if srcFs != nil {
		src, err = srcFs.Open(srcPath)
	} else {
		src, err = os.Open(srcPath)
	}
	if err != nil {
		return fmt.Errorf("florist.copyfile: open src file: %s", err)
	}
	defer src.Close()

	// if dstPath is an executable file and is running, then we will get back a
	// TXTBSY (text file busy).
	// The workaround is to unlink (or delete) the file beforehand.
	_ = os.Remove(dstPath) // useless to check for errors now

	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("florist.copyfile: open dst file: %s", err)
	}
	defer dst.Close()

	uid, _ := strconv.Atoi(owner.Uid)
	gid, _ := strconv.Atoi(owner.Gid)
	if err := os.Chown(dstPath, uid, gid); err != nil {
		return fmt.Errorf("florist.copyfile: %s", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("florist.copyfile: %s", err)
	}

	return nil
}

// Reads the whole file in memory, since it must do text template processing.
func copytemplate(
	srcFs fs.FS, srcPath string, dstPath string,
	mode os.FileMode, owner *user.User, tmplData any, delimL, delimR string,
) error {
	buf, err := fs.ReadFile(srcFs, srcPath)
	if err != nil {
		return fmt.Errorf("florist.copytemplate: %s", err)
	}

	tmpl, err := template.New(path.Base(srcPath)).Delims(delimL, delimR).Parse(string(buf))
	if err != nil {
		return fmt.Errorf("florist.copytemplate: %s", err)
	}
	// When looking up keys in a map, error out on missing key, as is the case for
	// struct missing field.
	tmpl.Option("missingkey=error")

	// if dstPath is an executable file and is running, then we will get back a
	// TXTBSY (text file busy).
	// The workaround is to unlink (or delete) the file beforehand.
	_ = os.Remove(dstPath) // useless to check for errors now

	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("florist.copytemplate: %s", err)
	}
	defer dst.Close()

	uid, _ := strconv.Atoi(owner.Uid)
	gid, _ := strconv.Atoi(owner.Gid)
	if err := os.Chown(dstPath, uid, gid); err != nil {
		return fmt.Errorf("florist.copytemplate: %s", err)
	}

	if err := tmpl.Execute(dst, tmplData); err != nil {
		return fmt.Errorf("florist.copytemplate: %s", err)
	}

	return nil
}
