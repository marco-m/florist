package florist

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"text/template"
)

func ListFs(fsys fs.FS) []string {
	if fsys == nil {
		return []string{"**** nil FS ****"}
	}

	var files []string

	fn := func(dePath string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !de.IsDir() {
			files = append(files, dePath)
		}
		return nil
	}

	if err := fs.WalkDir(fsys, ".", fn); err != nil {
		files = append(files, fmt.Sprintf("****%s****", err.Error()))
	}
	return files
}

func WriteFile(fname string, data string, perm os.FileMode, owner string) error {
	if err := Mkdir(path.Dir(fname), User().Username, 0700); err != nil {
		return fmt.Errorf("florist.WriteFile: %s", err)
	}

	Log().Debug("write-file", "name", fname)
	if err := os.WriteFile(fname, []byte(data), perm); err != nil {
		return fmt.Errorf("florist.WriteFile: %s", err)
	}
	// We call Chmod explicitly because os.WriteFile does _not_ changes the
	// mode _if_ the file already exists.
	if err := os.Chmod(fname, perm); err != nil {
		return fmt.Errorf("florist.WriteFile: %s", err)
	}

	if err := Chown(fname, owner); err != nil {
		return fmt.Errorf("florist.WriteFile: %s", err)
	}

	return nil
}

func Chown(fpath string, username string) error {
	ownerUser, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("florist.chown: %s", err)
	}
	uid, _ := strconv.Atoi(ownerUser.Uid)
	gid, _ := strconv.Atoi(ownerUser.Gid)
	if err := os.Chown(fpath, uid, gid); err != nil {
		return fmt.Errorf("florist.chown: %s", err)
	}
	return nil
}

// Mkdir creates directory dirPath, along with any necessary parents.
// The permission bits perm (before umask) are used for all directories that
// Mkdir creates. The owner of the intermediate parents will be the current
// user, while the owner of the last element of dirPath will be owner.
func Mkdir(dirPath string, owner string, perm fs.FileMode) error {
	log := Log().With("path", dirPath)
	log.Debug("mkdir", "owner", owner)

	ownerUser, err := user.Lookup(owner)
	if err != nil {
		return fmt.Errorf("florist.Mkdir: %s", err)
	}

	log.Debug("create-directory", "owner", owner, "perm", perm)
	if err := os.MkdirAll(dirPath, perm); err != nil {
		return fmt.Errorf("florist.Mkdir: %s", err)
	}

	uid, _ := strconv.Atoi(ownerUser.Uid)
	gid, _ := strconv.Atoi(ownerUser.Gid)
	log.Debug("chow-directory", "owner", owner)
	if err := os.Chown(dirPath, uid, gid); err != nil {
		return fmt.Errorf("florist.Mkdir: %s", err)
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
	mode os.FileMode, owner string,
) error {
	return copyfile(nil, srcPath, dstPath, mode, owner)
}

// CopyFileFs copies file srcPath to dstPath, with mode and owner. The source
// file resides in the srcFs filesystem (for example, via go:embed), while the
// destination file resides in the "real" filesystem.
// Notes:
// - If dstPath exists, it will be overwritten.
// - Setting an owner different that the current user requires elevated privileges.
func CopyFileFs(
	srcFs fs.FS, srcPath string, dstPath string,
	mode os.FileMode, owner string,
) error {
	return copyfile(srcFs, srcPath, dstPath, mode, owner)
}

// CopyTemplateFs copies file srcPath to dstPath, with mode and owner, performing
// Go text template processing based on the fields in tmplData. The source file resides
// in the srcFs filesystem (for example, via go:embed), while the destination file
// resides in the "real" filesystem.
// Notes:
//   - If dstPath exists, it will be overwritten.
//   - Setting an owner different that the current user requires elevated privileges.
//   - Template delimiters delimL, delimR allow to easily support templates that contain
//     themselves the default delimiters {{, }}. Passing the empty delimiters stand for
//     the default.
//func CopyTemplateFs(
//	srcFs fs.FS, srcPath string, dstPath string,
//	mode os.FileMode, owner string, tmplData any, delimL, delimR string,
//) error {
//	return copytemplate(srcFs, srcPath, dstPath, mode, owner, tmplData, delimL, delimR)
//}

func TemplateFromText(text string, tmplData any) (string, error) {
	Log().Debug("TemplateFromText")
	return rendertext(text, tmplData, "", "")
}

func TemplateFromFs(srcFs fs.FS, srcPath string, tmplData any) (string, error) {
	Log().Debug("TemplateFromFs", "file-name", srcPath)
	return rendertemplate(srcFs, srcPath, tmplData, "", "")
}

// TemplateFromFsWithDelims uses "<<", ">>" as template delimiters.
// This is useful to escape the default delimiters "{{", "}}" in the template.
func TemplateFromFsWithDelims(srcFs fs.FS, srcPath string, tmplData any) (string, error) {
	Log().Debug("TemplateFromFsWithDelims", "file-name", srcPath)
	return rendertemplate(srcFs, srcPath, tmplData, "<<", ">>")
}

// Reads the whole file in memory, since it must do text template processing.
func rendertemplate(
	srcFs fs.FS, srcPath string, tmplData any, delimL, delimR string,
) (string, error) {
	buf, err := fs.ReadFile(srcFs, srcPath)
	if err != nil {
		return "", fmt.Errorf("florist.rendertemplate: %s", err)
	}

	return rendertext(string(buf), tmplData, delimL, delimR)

	//// if dstPath is an executable file and is running, then we will get back a
	//// TXTBSY (text file busy).
	//// The workaround is to unlink (or delete) the file beforehand.
	//_ = os.Remove(dstPath) // useless to check for errors now
	//
	//dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	//if err != nil {
	//	return fmt.Errorf("florist.copytemplate: %s", err)
	//}
	//defer dst.Close()
	//
	//ownerUser, err := user.Lookup(owner)
	//if err != nil {
	//	return fmt.Errorf("florist.copytemplate: %s", err)
	//}
	//
	//uid, _ := strconv.Atoi(ownerUser.Uid)
	//gid, _ := strconv.Atoi(ownerUser.Gid)
	//if err := os.Chown(dstPath, uid, gid); err != nil {
	//	return fmt.Errorf("florist.copytemplate: %s", err)
	//}
	//
	//if err := tmpl.Execute(dst, tmplData); err != nil {
	//	return fmt.Errorf("florist.copytemplate: %s", err)
	//}
	//
	//return nil
}

func rendertext(text string, tmplData any, delimL, delimR string) (string, error) {
	tmpl := template.New("render-text").
		Delims(delimL, delimR).
		Option("missingkey=error")

	_, err := tmpl.Parse(text)
	if err != nil {
		return "", fmt.Errorf("florist.rendertext: %s", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tmplData); err != nil {
		return "", fmt.Errorf("florist.rendertext: %s", err)
	}

	return buf.String(), nil
}

// Does not read the whole file in memory.
func copyfile(
	srcFs fs.FS, srcPath string, dstPath string,
	mode os.FileMode, owner string,
) error {
	ownerUser, err := user.Lookup(owner)
	if err != nil {
		return fmt.Errorf("florist.copyfile: %s", err)
	}

	var src fs.File
	if srcFs != nil {
		src, err = srcFs.Open(srcPath)
	} else {
		src, err = os.Open(srcPath)
	}
	if err != nil {
		return fmt.Errorf("florist.copyfile: open src file: %s", err)
	}
	defer src.Close()

	if err := Mkdir(filepath.Dir(dstPath), owner, mode|0111); err != nil {
		return fmt.Errorf("florist.copyfile: %s", err)
	}

	// if dstPath is an executable file and is running, then we will get back a
	// TXTBSY (text file busy).
	// The workaround is to unlink (or delete) the file beforehand.
	_ = os.Remove(dstPath) // useless to check for errors now

	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("florist.copyfile: open dst file: %s", err)
	}
	defer dst.Close()

	uid, _ := strconv.Atoi(ownerUser.Uid)
	gid, _ := strconv.Atoi(ownerUser.Gid)
	if err := os.Chown(dstPath, uid, gid); err != nil {
		return fmt.Errorf("florist.copyfile: %s", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("florist.copyfile: %s", err)
	}

	return nil
}
