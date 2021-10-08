package florist

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"strconv"
)

// Copy copies file srcPath to dstPath, with mode and owner. The source and
// destination files reside in the "real" filesystem.
// Notes:
// - If dstPath exists, it will be overwritten.
// - Setting an owner different that the current user requires elevated privileges.
func Copy(srcPath string, dstPath string,
	mode os.FileMode, owner *user.User,
) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("florist.Copy: %s", err)
	}
	defer src.Close()

	return copy(src, dstPath, mode, owner)
}

// CopyFromFs copies file srcPath to dstPath, with mode and owner. The source
// file resides in the srcFs filesystem (for example, via go:embed), while the
// destination file resides in the "real" filesystem.
// Notes:
// - If dstPath exists, it will be overwritten.
// - Setting an owner different that the current user requires elevated privileges.
func CopyFromFs(srcFs fs.FS, srcPath string,
	dstPath string,
	mode os.FileMode, owner *user.User,
) error {
	src, err := srcFs.Open(srcPath)
	if err != nil {
		return fmt.Errorf("florist.CopyFromFS: %s", err)
	}
	defer src.Close()

	return copy(src, dstPath, mode, owner)
}

func copy(src io.Reader, dstPath string,
	mode os.FileMode, owner *user.User,
) error {
	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("florist.copy: %s", err)
	}
	defer dst.Close()

	uid, _ := strconv.Atoi(owner.Uid)
	gid, _ := strconv.Atoi(owner.Gid)
	if err := os.Chown(dstPath, uid, gid); err != nil {
		return fmt.Errorf("florist.copy: %s", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("florist.copy: %s", err)
	}

	return nil
}
