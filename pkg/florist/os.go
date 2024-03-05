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

// WriteFile writes data to fname and sets the mode and owner of fname.
// If also creates any missing directories in the path, if any.
func WriteFile(fname string, data string, mode os.FileMode, owner string) error {
	if err := os.MkdirAll(path.Dir(fname), 0700); err != nil {
		return fmt.Errorf("florist.WriteFile: %s", err)
	}

	Log().Debug("write-file", "name", fname)
	if err := os.WriteFile(fname, []byte(data), mode); err != nil {
		return fmt.Errorf("florist.WriteFile: %s", err)
	}
	// We call Chmod explicitly because os.WriteFile does _not_ changes the
	// mode _if_ the file already exists.
	if err := os.Chmod(fname, mode); err != nil {
		return fmt.Errorf("florist.WriteFile: %s", err)
	}

	if err := Chown(fname, owner); err != nil {
		return fmt.Errorf("florist.WriteFile: %s", err)
	}

	return nil
}

// Chown sets the owner of fpath to username.
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

// TemplateFromText renders the template text with tmplData.
func TemplateFromText(text string, tmplData any) (string, error) {
	Log().Debug("TemplateFromText")
	return renderText(text, tmplData, "", "")
}

// TemplateFromFs reads file srcPath in filesystem srcFs and renders its contents
// // as a template with data tmplData.
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

	return renderText(string(buf), tmplData, delimL, delimR)

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

// renderText renders the template text with data tmplData, with delimL and delimR
// as template delimiters. If delimL and delimR are empty, the default delimiters
// "{{" and "}}" will be used.
func renderText(text string, tmplData any, delimL, delimR string) (string, error) {
	tmpl := template.New("render-text").
		Delims(delimL, delimR).
		Option("missingkey=error")

	_, err := tmpl.Parse(text)
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

	if err := os.MkdirAll(filepath.Dir(dstPath), mode|0111); err != nil {
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
