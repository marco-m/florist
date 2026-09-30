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

func (cmd *configureCmd) Run(app App) error {
	run := func() error {
		config, err := NewConfig(cmd.Settings)
		if err != nil {
			app.prov.errs = append(app.prov.errs, err)
		}

		app.log.Info("preconfigure-running")
		var bag any
		if bag, err = app.opts.PreConfigureFn(app.prov, config); err != nil {
			app.prov.errs = append(app.prov.errs, fmt.Errorf("preconfigure: %s", err))
		}

		app.log.Info("configuring-each-flower", "flowers-count", len(app.prov.flowers),
			"flowers", app.prov.ordered)

		for _, k := range app.prov.ordered {
			fl := app.prov.flowers[k]
			app.log.Info("configuring", "flower", fl.String())
			if err := fl.Init(); err != nil {
				app.prov.errs = append(app.prov.errs, fmt.Errorf("flower init: %s", err))
			}
			if err := fl.Configure(app.opts.Seeds); err != nil {
				app.prov.errs = append(app.prov.errs, fmt.Errorf("flower configure: %s", err))
			}
		}

		if cfgErr := config.Errors(); cfgErr != nil {
			app.prov.errs = append(app.prov.errs, cfgErr)
		}
		if app.opts.PostConfigureFn != nil {
			app.log.Info("postconfigure-running")
			if err := app.opts.PostConfigureFn(app.prov, config, bag); err != nil {
				app.prov.errs = append(app.prov.errs, fmt.Errorf("postconfigure: %s", err))
			}
		} else {
			app.log.Info("postconfigure-nothing-to-run")
		}

		status := "✅ success"
		if len(app.prov.errs) > 0 {
			status = "❌ failure"
		}
		if err := customizeMotd("configured", status, app.opts.RootDir); err != nil {
			app.prov.errs = append(app.prov.errs, err)
		}

		if err := JoinErrors(app.prov.errs...); err != nil {
			return fmt.Errorf("configure: %s", err)
		}
		return nil
	}

	return timelog(run, app)
}
