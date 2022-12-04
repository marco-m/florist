// Package copyfiles is a very simple example of a flower.
package copyfiles

import (
	"fmt"
	"io/fs"
	"os/user"
	"path"

	"github.com/marco-m/florist"
)

const (
	// Another option would be to make the DstDir a parameter of the Flower.
	DstDir = "/tmp/example-dst"
)

type Flower struct {
	FilesFS  fs.FS
	SrcFiles []string
}

func (fl *Flower) String() string {
	// When writing your own flower, replace "example" with the name of your project.
	return "example.flower.copyfiles"
}

func (fl *Flower) Description() string {
	return "copy files from an embed.FS to the real filesystem"
}

func (fl *Flower) Init() error {
	return nil
}

func (fl *Flower) Install() error {
	log := florist.Log.ResetNamed(fl.String())
	log.Info("begin")
	defer log.Info("end")

	curUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("Create dst dir", "dir", DstDir)
	if err := florist.Mkdir(DstDir, curUser, 0755); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Debug("installing files", "total", len(fl.SrcFiles))
	for _, cfg := range fl.SrcFiles {
		src := cfg
		dst := path.Join(DstDir, cfg)
		log.Info("Install file", "dst", dst)
		if err := florist.CopyFileFromFs(fl.FilesFS, src, dst, 0644, curUser); err != nil {
			return fmt.Errorf("%s: %s", log.Name(), err)
		}
	}

	return nil
}

func (fl *Flower) Configure(rawCfg []byte) error {
	return nil
}
