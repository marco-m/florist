package florist

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"text/template"
)

// ListFs returns a list of the files (not directories) in fsys.
// In case of error, it encodes the error in a file name in the list.
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

// Return true if file 'fpath' exists.
// WARNING Checking for file existence is racy and in certain cases can lead
// to security vulnerabilities. Think twice before using this. In the majority
// of cases, you can simply skip the existence check, since the next operation
// will fail in any case if the file doesn't exist.
//
// Explanation of the TOCTOU vulnerability:
// https://wiki.sei.cmu.edu/confluence/display/c/FIO45-C.+Avoid+TOCTOU+race+conditions+while+accessing+files
func FileExists(fpath string) (bool, error) {
	_, err := os.Stat(fpath)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// WriteFile writes 'data' to 'fname' and sets 'mode', 'owner' and 'group' of
// 'fname'. WriteFile will fail if the base directory of 'fname' doesn't exist.
// This is done on purpose to ensure that the caller addresses explicitly the
// ownership and permissions of the path segments containing 'fname'.
func WriteFile(fname string, data string,
	mode os.FileMode, owner string, group string,
) error {
	Log().Debug("WriteFile", "name", fname)
	if err := os.WriteFile(fname, []byte(data), mode); err != nil {
		return fmt.Errorf("florist.WriteFile: %s", err)
	}
	// We call Chmod explicitly because os.WriteFile does _not_ change the
	// mode _if_ the file already exists.
	if err := os.Chmod(fname, mode); err != nil {
		return fmt.Errorf("florist.WriteFile: %s", err)
	}
	if err := Chown(fname, owner); err != nil {
		return fmt.Errorf("florist.WriteFile: %s", err)
	}
	if err := Chgrp(fname, group); err != nil {
		return fmt.Errorf("florist.WriteFile: %s", err)
	}

	return nil
}

// Chown sets the owner of 'fpath' to the user ID and primary group ID of 'username'.
// See also [Chgrp].
func Chown(fpath string, username string) error {
	theUser, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("florist.chown: %s", err)
	}
	uid, err := strconv.Atoi(theUser.Uid)
	if err != nil {
		return fmt.Errorf("florist.chown: %s", err)
	}
	gid, err := strconv.Atoi(theUser.Gid)
	if err != nil {
		return fmt.Errorf("florist.chown: %s", err)
	}
	if err := os.Chown(fpath, uid, gid); err != nil {
		return fmt.Errorf("florist.chown: %s", err)
	}
	return nil
}

// Chgrp sets the group of 'fpath' to the group ID of 'groupname'.
// See also [Chown].
func Chgrp(fpath string, groupname string) error {
	theGroup, err := user.LookupGroup(groupname)
	if err != nil {
		return fmt.Errorf("florist.chgrp: %s", err)
	}
	uid := -1 // Meaning: do not change the UID.
	gid, err := strconv.Atoi(theGroup.Gid)
	if err != nil {
		return fmt.Errorf("florist.chgrp: %s", err)
	}
	if err := os.Chown(fpath, uid, gid); err != nil {
		return fmt.Errorf("florist.chgrp: %s", err)
	}
	return nil
}

// Mkdir creates the directory 'fpath' with absolute permissions 'perm' (does
// not considers umask) and sets its user to 'owner' and its group to 'group'.
// If the directory 'fpath' already exists, Mkdir does not consider this as an
// error and proceeds, eventually overriding the previous ownership and
// permissions.
func Mkdir(fpath string, perm os.FileMode, owner string, group string) error {
	err := os.Mkdir(fpath, perm)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("florist.Mkdir: %s", err)
	}

	theUser, err := user.Lookup(owner)
	if err != nil {
		return fmt.Errorf("florist.Mkdir: %s", err)
	}
	uid, err := strconv.Atoi(theUser.Uid)
	if err != nil {
		return fmt.Errorf("florist.Mkdir: %s", err)
	}
	theGroup, err := user.LookupGroup(group)
	if err != nil {
		return fmt.Errorf("florist.Mkdir: %s", err)
	}
	gid, err := strconv.Atoi(theGroup.Gid)
	if err != nil {
		return fmt.Errorf("florist.Mkdir: %s", err)
	}

	if err := os.Chmod(fpath, perm); err != nil {
		return fmt.Errorf("florist.Mkdir: %s", err)
	}
	if err := os.Chown(fpath, uid, gid); err != nil {
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
	return copyFile(nil, srcPath, dstPath, mode, owner)
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
	return copyFile(srcFs, srcPath, dstPath, mode, owner)
}

// TemplateFromText renders the template 'tmplText' with data 'tmplData'.
// Parameter 'tmplName' is used for debugging purposes, a typical example is
// the template file name.
func TemplateFromText(tmplText string, tmplData any, tmplName string) (string, error) {
	Log().Debug("TemplateFromText")
	return renderText(tmplText, tmplData, tmplName, "", "")
}

// TemplateFromFs reads file srcPath in filesystem srcFs and renders its contents
// as a template with data tmplData.
func TemplateFromFs(srcFs fs.FS, srcPath string, tmplData any) (string, error) {
	Log().Debug("TemplateFromFs", "file-name", srcPath)
	return renderTemplate(srcFs, srcPath, tmplData, "", "")
}

// TemplateFromFsWithDelims reads file srcPath in filesystem srcFs and renders its
// contents as a template with data tmplData, with "<<", ">>" as template delimiters.
// This is useful to escape the default delimiters "{{", "}}" in the template.
func TemplateFromFsWithDelims(srcFs fs.FS, srcPath string, tmplData any) (string, error) {
	Log().Debug("TemplateFromFsWithDelims", "file-name", srcPath)
	return renderTemplate(srcFs, srcPath, tmplData, "<<", ">>")
}

// RenderFile renders the template in 'srcFs/srcFile' with the data in 'fl' and
// writes it to 'dstFile', with permissions 'mode' and 'owner' and 'group'.
// The template must use the delimiters '<<' and '>>'.
func RenderFile(log *slog.Logger, srcFs fs.FS, srcFile string, dstFile string,
	mode fs.FileMode, owner string, group string, fl any,
) error {
	log.Info("rendering file", "src", srcFile, "dst", dstFile)
	rendered, err := TemplateFromFsWithDelims(srcFs, srcFile, fl)
	if err != nil {
		return err
	}
	if err := WriteFile(dstFile, rendered, mode, owner, group); err != nil {
		return err
	}
	return nil
}

// renderTemplate reads file srcPath in filesystem srcFs and renders its contents
// as a template with data tmplData, with delimL and delimR as template delimiters.
// If delimL and delimR are empty, the default delimiters "{{" and "}}" will be used.
func renderTemplate(
	srcFs fs.FS, srcPath string, tmplData any, delimL, delimR string,
) (string, error) {
	buf, err := fs.ReadFile(srcFs, srcPath)
	if err != nil {
		return "", fmt.Errorf("florist.renderTemplate: %s", err)
	}

	return renderText(string(buf), tmplData, srcPath, delimL, delimR)

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

// renderText renders the template 'tmplText' with data 'tmplData'.
// Parameter 'tmplName' is used for debugging purposes, a typical example is
// the template file name.
// Parameters 'delimL' and 'delimR' as template delimiters. If they are empty,
// then the default delimiters "{{" and "}}" will be used.
func renderText(tmplText string, tmplData any, tmplName string,
	delimL, delimR string,
) (string, error) {
	tmpl := template.New(tmplName).
		Delims(delimL, delimR).
		Option("missingkey=error")

	_, err := tmpl.Parse(tmplText)
	if err != nil {
		return "", fmt.Errorf("florist.renderText: %s", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tmplData); err != nil {
		return "", fmt.Errorf("florist.renderText: %s", err)
	}

	return buf.String(), nil
}

// Does not read the whole file in memory.
func copyFile(
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

	if err := os.MkdirAll(filepath.Dir(dstPath), mode|0o111); err != nil {
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
