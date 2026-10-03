package mint

import (
	"fmt"
	"path/filepath"

	"github.com/marco-m/florist/pkg/florist"
)

const Name = "mint"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	gdn *florist.Garden

	DstDir string
}

type Conf struct {
	Aroma string
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

func (fl *Flower) Init(gdn *florist.Garden) error {
	fl.gdn = gdn

	// Defaults.
	if fl.DstDir == "" {
		fl.DstDir = filepath.Join(gdn.WorkDir(), "daisy")
	}
	if fl.Aroma == "" {
		fl.Aroma = "PepperMint"
	}

	return nil
}

func (fl *Flower) Install() error {
	// log := slog.With("flower", Name + ".install")
	text := `DstDir: {{.DstDir}}\n`
	rendered, err := florist.TemplateFromText(text, fl, "template-name")
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	dstPath := filepath.Join(fl.Inst.DstDir, "install.txt")
	if err := florist.WriteFile(dstPath, rendered, 0o600,
		fl.gdn.User(), fl.gdn.Group()); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	return nil
}

func (fl *Flower) Configure() error {
	// log := slog.With("flower", Name + ".configure")

	text := `DstDir: {{.DstDir}}
Aroma: {{.Aroma}}
`
	rendered, err := florist.TemplateFromText(text, fl, "template-name")
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	dstPath := filepath.Join(fl.Inst.DstDir, "configure.txt")
	if err := florist.WriteFile(dstPath, rendered, 0o600,
		fl.gdn.User(), fl.gdn.Group()); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	return nil
}
