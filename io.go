package florist

import (
	"io"
	"io/fs"
	"os"
	"os/user"
	"strconv"
)

// Copy file srcPath to dstPath, with mode and owner. The source file resides in
// the srcFs filesystem (for example, via go:embed), while the destination file
// resides in the "real" filesystem.
// Notes:
// - If dstPath exists, it will be overwritten.
// - To use the "real" filesystem as source, pass os.DirFS("/").
// - Setting an owner different that the current user requires elevated privileges.
func Copy(srcFs fs.FS, srcPath string,
	dstPath string,
	mode os.FileMode, owner *user.User,
) error {
	src, err := srcFs.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dst.Close()

	uid, _ := strconv.Atoi(owner.Uid)
	gid, _ := strconv.Atoi(owner.Gid)
	if err := os.Chown(dstPath, uid, gid); err != nil {
		return err
	}

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}
