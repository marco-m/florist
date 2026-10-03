package florist

import (
	"fmt"

	"github.com/marco-m/clim"
)

type installCmd struct{}

func newInstallCmd(parent *clim.CLI) (*installCmd, error) {
	installCmd := installCmd{}

	_, err := clim.NewSub(parent, "install", "install the flowers")
	return &installCmd, err
}

func (cmd *installCmd) Run(gdn *Garden) error {
	run := func() error {
		gdn.log.Info("installing", "flowers-count", len(gdn.prov.flowers),
			"flowers", gdn.prov.ordered)

		for _, k := range gdn.prov.ordered {
			fl := gdn.prov.flowers[k]
			gdn.log.Info("installing", "flower", fl.String())
			if err := fl.Init(gdn); err != nil {
				return fmt.Errorf("install: %s", err)
			}
			if err := fl.Install(); err != nil {
				return err
			}
		}

		status := "✅  success"
		return customizeMotd("installed", status, gdn.opts.RootDir)
	}

	return timelog(run, gdn)
}
