package florist

import (
	"fmt"

	"github.com/marco-m/clim"
)

type configureCmd struct {
	Settings string
}

func newConfigureCmd(parent *clim.CLI[App]) error {
	configureCmd := configureCmd{}

	cli, err := clim.NewSub(parent, "configure", "configure the flowers", configureCmd.Run)
	if err != nil {
		return err
	}

	if err := cli.AddFlags(&clim.Flag{
		Value: clim.String(&configureCmd.Settings, "/opt/florist/config.json"),
		Long:  "settings", Help: "Settings file (JSON)",
	}); err != nil {
		return err
	}

	return nil
}

func (cmd *configureCmd) Run(app App) error {
	run := func() error {
		config, err := NewConfig(cmd.Settings)
		if err != nil {
			return fmt.Errorf("configure: %s", err)
		}

		app.log.Info("running-preConfigure")
		var bag any
		if bag, err = app.opts.PreConfigureFn(app.prov, config); err != nil {
			return fmt.Errorf("configure: preConfigure: %s", err)
		}
		if err := config.Errors(); err != nil {
			return fmt.Errorf("configure: %s", err)
		}

		app.log.Info("configuring-each-flower", "flowers-count", len(app.prov.flowers),
			"flowers", app.prov.ordered)

		for _, k := range app.prov.ordered {
			fl := app.prov.flowers[k]
			app.log.Info("configuring", "flower", fl.String())
			if err := fl.Init(); err != nil {
				return fmt.Errorf("configure: %s", err)
			}
			if err := fl.Configure(); err != nil {
				return fmt.Errorf("configure: %s", err)
			}
		}

		if err := customizeMotd("configured", app.prov.rootDir); err != nil {
			return fmt.Errorf("configure: %s", err)
		}

		if app.opts.PostConfigureFn != nil {
			app.log.Info("postConfigure-running")
			if err := app.opts.PostConfigureFn(app.prov, config, bag); err != nil {
				return fmt.Errorf("configure: postConfigure: %s", err)
			}
		} else {
			app.log.Info("postConfigure-nothing-to-run")
		}
		return nil
	}

	return timelog(run, app)
}
