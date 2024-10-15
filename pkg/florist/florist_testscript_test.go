package florist_test

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/marco-m/florist/pkg/florist"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"provisioner-1": func() int {
			return florist.MainInt(&florist.Options{
				SetupFn:     setup,
				ConfigureFn: configure,
			},
			)
		},
	}))
}

// Look at testdata/*.txt and testdata/*.txtar for the actual tests.
func TestScriptFlorist(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}

func setup(prov *florist.Provisioner) error {
	prov.UseWorkdir() // FIXME what is this???
	err := prov.AddFlowers(
		&testFlower{
			Inst: Inst{FieldInst: "from-installed"},
		},
	)
	return err
}

func configure(prov *florist.Provisioner, config *florist.Config) error {
	prov.Flowers()["testFlower"].(*testFlower).Conf = Conf{
		FieldConf: "from-configured",
	}
	return florist.JoinErrors(config.Errors())
}

type testFlower struct {
	Inst
	Conf
}

type Inst struct {
	FieldInst string
}

type Conf struct {
	FieldConf string
}

func (fl testFlower) String() string {
	return "testFlower"
}

func (fl testFlower) Description() string {
	return "description of testFllower"
}

func (fl testFlower) Embedded() []string {
	return []string{"one", "two"}
}

func (fl testFlower) Init() error {
	return nil
}

func (fl testFlower) Install() error {
	dstPath := "installed.txt"
	rendered, err := florist.TemplateFromText("{{.FieldInst}}\n", fl)
	if err != nil {
		return err
	}
	username := florist.User().Username
	if err := florist.WriteFile(dstPath, rendered, 0o600, username); err != nil {
		return err
	}
	return nil
}

func (fl testFlower) Configure() error {
	dstPath := "configured.txt"
	rendered, err := florist.TemplateFromText("{{.FieldInst}} {{.FieldConf}}\n", fl)
	if err != nil {
		return err
	}
	username := florist.User().Username
	if err := florist.WriteFile(dstPath, rendered, 0o600, username); err != nil {
		return err
	}
	return nil
}
