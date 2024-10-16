package mint

import (
	"fmt"
	"path/filepath"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/pkg/florist"
)

const Name = "mint"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	DstDir string `default:"/tmp/mint"`
}

type Conf struct {
	Aroma string `default:"PepperMint"`
}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "a mint flower"
}

func (fl *Flower) Embedded() []string {
	return nil
}

func (fl *Flower) Init() error {
	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	return nil
}

func (fl *Flower) Install() error {
	// log := florist.Log().With("flower", Name + ".install")

	text := `DstDir: {{.DstDir}}\n`
	rendered, err := florist.TemplateFromText(text, fl, "template-name")
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	username := florist.User().Username
	dstPath := filepath.Join(fl.Inst.DstDir, "install.txt")
	if err := florist.WriteFile(dstPath, rendered, 0o600, username); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	return nil
}

func (fl *Flower) Configure() error {
	// log := florist.Log().With("flower", Name + ".configure")

	text := `DstDir: {{.DstDir}}
Aroma: {{.Aroma}}
`
	rendered, err := florist.TemplateFromText(text, fl, "template-name")
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	username := florist.User().Username
	dstPath := filepath.Join(fl.Inst.DstDir, "configure.txt")
	if err := florist.WriteFile(dstPath, rendered, 0o600, username); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	return nil
}
