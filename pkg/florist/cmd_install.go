package florist

import (
	"fmt"

	"github.com/marco-m/clim"
)

type installCmd struct{}

func newInstallCmd(parent *clim.CLI[App]) error {
	installCmd := installCmd{}

	_, err := clim.NewSub(parent, "install", "install the flowers", installCmd.Run)
	return err
}

func (cmd *installCmd) Run(app App) error {
	run := func() error {
		app.log.Info("installing", "flowers-count", len(app.prov.flowers),
			"flowers", app.prov.ordered)

		for _, k := range app.prov.ordered {
			fl := app.prov.flowers[k]
			app.log.Info("installing", "flower", fl.String())
			if err := fl.Init(); err != nil {
				return fmt.Errorf("install: %s", err)
			}
			if err := fl.Install(); err != nil {
				return err
			}
		}
		return customizeMotd("installed", app.prov.rootDir)
	}

	return timelog(run, app)
}
