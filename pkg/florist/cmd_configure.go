package florist

import (
	"fmt"
	"path/filepath"

	"github.com/marco-m/clim"
)

type configureCmd struct {
	Settings string
}

func newConfigureCmd(parent *clim.CLI) (*configureCmd, error) {
	configureCmd := configureCmd{}

	cli, err := clim.NewSub(parent, "configure", "configure the flowers")
	if err != nil {
		return nil, err
	}

	if err := cli.AddFlags(&clim.Flag{
		Value: clim.String(&configureCmd.Settings, filepath.Join(HomeDir, "config.json")),
		Long:  "settings", Help: "Settings file (JSON)",
	}); err != nil {
		return nil, err
	}

	return &configureCmd, nil
}

func (cmd *configureCmd) Run(gdn *Garden) error {
	run := func() error {
		config, err := NewConfig(cmd.Settings)
		if err != nil {
			gdn.prov.errs = append(gdn.prov.errs, err)
		}

		gdn.log.Info("preconfigure-running")
		var bag any
		if bag, err = gdn.opts.PreConfigureFn(gdn.prov, config); err != nil {
			gdn.prov.errs = append(gdn.prov.errs, fmt.Errorf("preconfigure: %s", err))
		}

		gdn.log.Info("configuring-each-flower", "flowers-count", len(gdn.prov.flowers),
			"flowers", gdn.prov.ordered)

		for _, k := range gdn.prov.ordered {
			fl := gdn.prov.flowers[k]
			gdn.log.Info("configuring", "flower", fl.String())
			if err := fl.Init(gdn); err != nil {
				gdn.prov.errs = append(gdn.prov.errs, fmt.Errorf("flower init: %s", err))
			}
			if err := fl.Configure(); err != nil {
				gdn.prov.errs = append(gdn.prov.errs, fmt.Errorf("flower configure: %s", err))
			}
		}

		if cfgErr := config.Errors(); cfgErr != nil {
			gdn.prov.errs = append(gdn.prov.errs, cfgErr)
		}
		if gdn.opts.PostConfigureFn != nil {
			gdn.log.Info("postconfigure-running")
			if err := gdn.opts.PostConfigureFn(gdn.prov, config, bag); err != nil {
				gdn.prov.errs = append(gdn.prov.errs, fmt.Errorf("postconfigure: %s", err))
			}
		} else {
			gdn.log.Info("postconfigure-nothing-to-run")
		}

		status := "✅ success"
		if len(gdn.prov.errs) > 0 {
			status = "❌ failure"
		}
		if err := customizeMotd("configured", status, gdn.opts.RootDir); err != nil {
			gdn.prov.errs = append(gdn.prov.errs, err)
		}

		if err := JoinErrors(gdn.prov.errs...); err != nil {
			return fmt.Errorf("configure: %s", err)
		}
		return nil
	}

	return timelog(run, gdn)
}
