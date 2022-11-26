package florist

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"strconv"
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
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("florist.CopyFile: %s", err)
	}
	defer src.Close()

	return copyfile(src, dstPath, mode, owner)
}

// CopyFileFromFs copies file srcPath to dstPath, with mode and owner. The source
// file resides in the srcFs filesystem (for example, via go:embed), while the
// destination file resides in the "real" filesystem.
// Notes:
// - If dstPath exists, it will be overwritten.
// - Setting an owner different that the current user requires elevated privileges.
func CopyFileFromFs(srcFs fs.FS,
	srcPath string, dstPath string,
	mode os.FileMode, owner *user.User,
) error {
	src, err := srcFs.Open(srcPath)
	if err != nil {
		return fmt.Errorf("florist.CopyFromFS: %s", err)
	}
	defer src.Close()

	return copyfile(src, dstPath, mode, owner)
}

func copyfile(src io.Reader, dstPath string,
	mode os.FileMode, owner *user.User,
) error {
	// if dstPath is an executable file and is running, then we will get back a
	// TXTBSY (text file busy).
	// The workaround is to unlink (or delete) the file beforehand.
	_ = os.Remove(dstPath) // useless to check for errors now

	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("florist.filecopy: %s", err)
	}
	defer dst.Close()

	uid, _ := strconv.Atoi(owner.Uid)
	gid, _ := strconv.Atoi(owner.Gid)
	if err := os.Chown(dstPath, uid, gid); err != nil {
		return fmt.Errorf("florist.filecopy: %s", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("florist.filecopy: %s", err)
	}

	return nil
}
