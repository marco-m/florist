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

	uid, _ := strconv.Atoi(owner.Uid)
	gid, _ := strconv.Atoi(owner.Gid)

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
	if err := os.Mkdir(path, perm); err != nil {
		return fmt.Errorf("florist.Mkdir: %s", err)
	}
	log.Debug("directory created", "perm", perm)
	if err := os.Chown(path, uid, gid); err != nil {
		return fmt.Errorf("florist.Mkdir: %s", err)
	}
	log.Debug("directory chown", "owner", owner.Username)
	return nil
}

// CopyFile copies file srcPath to dstPath, with mode and owner. The source and
// destination files reside in the "real" filesystem.
// Notes:
// - If dstPath exists, it will be overwritten.
// - Setting an owner different that the current user requires elevated privileges.
func CopyFile(srcPath string, dstPath string,
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
func CopyFileFromFs(srcFs fs.FS, srcPath string, dstPath string,
	mode os.FileMode, owner *user.User,
) error {
	return copyfile(srcFs, srcPath, dstPath, mode, owner)
}

// CopyFileTemplate copies file srcPath to dstPath, with mode and owner, performing Go
// text template processing based on the fields in tmplData. The source and destination
// files reside in the "real" filesystem.
// Notes:
// - If dstPath exists, it will be overwritten.
// - Setting an owner different that the current user requires elevated privileges.
func CopyFileTemplate(srcPath string, dstPath string,
	mode os.FileMode, owner *user.User, tmplData any,
) error {
	return copyfiletmpl(nil, srcPath, dstPath, mode, owner, tmplData)
}

// CopyFileTemplateFromFs copies file srcPath to dstPath, with mode and owner, performing
// Go text template processing based on the fields in tmplData. The source file resides
// in the srcFs filesystem (for example, via go:embed), while the destination file
// resides in the "real" filesystem.
// Notes:
// - If dstPath exists, it will be overwritten.
// - Setting an owner different that the current user requires elevated privileges.
func CopyFileTemplateFromFs(srcFs fs.FS, srcPath string, dstPath string,
	mode os.FileMode, owner *user.User, tmplData any,
) error {
	return copyfiletmpl(srcFs, srcPath, dstPath, mode, owner, tmplData)
}

// Does not read the whole file in memory.
func copyfile(srcFs fs.FS, srcPath string, dstPath string,
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
func copyfiletmpl(srcFs fs.FS, srcPath string, dstPath string,
	mode os.FileMode, owner *user.User, tmplData any,
) error {
	var buf []byte
	var err error
	if srcFs != nil {
		buf, err = fs.ReadFile(srcFs, srcPath)
	} else {
		buf, err = os.ReadFile(srcPath)
	}
	if err != nil {
		return fmt.Errorf("florist.copyfile: %s", err)
	}

	tmpl, err := template.New(path.Base(srcPath)).Parse(string(buf))
	if err != nil {
		return fmt.Errorf("florist.copyfile: %s", err)
	}

	// if dstPath is an executable file and is running, then we will get back a
	// TXTBSY (text file busy).
	// The workaround is to unlink (or delete) the file beforehand.
	_ = os.Remove(dstPath) // useless to check for errors now

	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("florist.copyfile: %s", err)
	}
	defer dst.Close()

	uid, _ := strconv.Atoi(owner.Uid)
	gid, _ := strconv.Atoi(owner.Gid)
	if err := os.Chown(dstPath, uid, gid); err != nil {
		return fmt.Errorf("florist.copyfile: %s", err)
	}

	if err := tmpl.Execute(dst, tmplData); err != nil {
		return fmt.Errorf("florist.copyfile: %s", err)
	}

	return nil
}
