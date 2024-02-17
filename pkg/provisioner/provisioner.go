// Package provisioner provides helpers to write your own florist provisioner.
package provisioner

import (
	"fmt"
	stdlog "log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/hashicorp/go-hclog"
	"tailscale.com/util/multierr"

	"github.com/marco-m/florist/pkg/florist"
)

type Provisioner struct {
	log           hclog.Logger
	cacheValidity time.Duration
	flowers       map[string]florist.Flower
	ordered       []string
	root          string
	errs          []string
}

func New(log hclog.Logger, cacheValidity time.Duration,
) (*Provisioner, error) {
	florist.SetLogger(log)

	stdlog.SetOutput(log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	stdlog.SetPrefix("")
	stdlog.SetFlags(0)

	return &Provisioner{
		log:           log.Named("provisioner"),
		cacheValidity: cacheValidity,
		flowers:       make(map[string]florist.Flower),
	}, nil
}

// FIXME what is this doing????
func (prov *Provisioner) UseWorkdir() {
	prov.root = florist.WorkDir
}

// Flowers returns
func (prov *Provisioner) Flowers() map[string]florist.Flower {
	return prov.flowers
}

func (prov *Provisioner) AddFlowers(flowers ...florist.Flower) error {
	if len(prov.flowers) > 0 {
		return fmt.Errorf("Provisioner.AddFlowers: cannot call more than once")
	}
	for i, flower := range flowers {
		if flower.String() == "" {
			return fmt.Errorf("Provisioner.AddFlowers: flower %d has empty name", i)
		}
		if flower.Description() == "" {
			return fmt.Errorf("Provisioner.AddFlowers: flower %s has empty description",
				flower)
		}
		if _, found := prov.flowers[flower.String()]; found {
			return fmt.Errorf("Provisioner.AddFlowers: flower with same name already exists: %s", flower)
		}
		prov.ordered = append(prov.ordered, flower.String())
		prov.flowers[flower.String()] = flower
	}
	return nil
}

type cli struct {
	Install   InstallCmd   `cmd:"" help:"install the ${role} role"`
	Configure ConfigureCmd `cmd:"" help:"configure the ${role} role"`
	List      ListCmd      `cmd:"" help:"list the flowers and their files"`
}

type InstallCmd struct{}

func (cmd *InstallCmd) Run(ctx *context) error {
	if err := florist.Init(); err != nil {
		return err
	}
	prov, err := ctx.setup(ctx.log)
	if err != nil {
		return fmt.Errorf("install: %s", err)
	}

	prov.log.Info("installing", "flower-size", len(prov.flowers),
		"flowers", prov.ordered)

	for _, k := range prov.ordered {
		fl := prov.flowers[k]
		prov.log.Info("Installing", "flower", fl.String())
		if err := fl.Init(); err != nil {
			return fmt.Errorf("install: %s", err)
		}
		if err := fl.Install(); err != nil {
			return err
		}
	}

	return customizeMotd(prov.log, "installed", prov.root)
}

type ConfigureCmd struct {
	Settings []string `help:"Settings and secrets file(s) (JSON)"`
}

func (cmd *ConfigureCmd) Run(ctx *context) error {
	if err := florist.Init(); err != nil {
		return fmt.Errorf("configure: %s", err)
	}

	config, err := florist.NewConfig(cmd.Settings)
	if err != nil {
		return fmt.Errorf("configure: %s", err)
	}

	prov, err := ctx.setup(ctx.log)
	if err != nil {
		return fmt.Errorf("configure: %s", err)
	}
	if err := ctx.configure(prov, config); err != nil {
		return fmt.Errorf("configure: %s", err)
	}
	if err := config.Errors(); err != nil {
		return fmt.Errorf("configure: %s", err)
	}

	prov.log.Info("configuring", "flowers-size", len(prov.flowers),
		"flowers", prov.ordered)

	for _, k := range prov.ordered {
		fl := prov.flowers[k]
		prov.log.Info("configuring", "flower", fl.String())
		if err := fl.Init(); err != nil {
			return fmt.Errorf("configure: %s", err)
		}
		if err := fl.Configure(); err != nil {
			return fmt.Errorf("configure: %s", err)
		}
	}

	if err := customizeMotd(prov.log, "configured", prov.root); err != nil {
		return fmt.Errorf("configure: %s", err)
	}
	return nil
}

type ListCmd struct {
}

func (cmd *ListCmd) Run(ctx *context) error {
	prov, err := ctx.setup(ctx.log)
	if err != nil {
		return fmt.Errorf("list: %s", err)
	}

	for _, k := range prov.ordered {
		v := prov.flowers[k]
		if err := v.Init(); err != nil {
			return err
		}
		fmt.Printf("%s -- %s\n", v, v.Description())
		for _, fi := range v.Embedded() {
			fmt.Printf("  %s\n", fi)
		}
	}
	return nil
}

type context struct {
	log       hclog.Logger
	setup     SetupFn
	configure ConfigureFn
}

// SetupFn is the function signature to pass to [Main].
type SetupFn func(log hclog.Logger) (*Provisioner, error)

// ConfigureFn is the function signature to pass to [Main].
type ConfigureFn func(prov *Provisioner, config *florist.Config) error

// Main is a ready-made function for the main() of your installer.
// Usage:
//
//	func main() {
//		log := florist.NewLogger("example")
//		os.Exit(installer.Main(log, os.Args, setup, configure))
//	}
//
// If you don't want to use it, just copy it and modify as you see fit.
func Main(log hclog.Logger, setup SetupFn, configure ConfigureFn) int {
	var cli cli
	parser, err := kong.New(&cli,
		kong.Description("🌼 florist 🌺 - a simple provisioner"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		},
		),
		kong.Vars{
			"role": filepath.Base(os.Args[0]),
		},
	)
	if err != nil {
		panic(err)
	}
	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	start := time.Now()
	log.Info("starting", "invocation", os.Args)
	err = ctx.Run(&context{log: log, setup: setup, configure: configure})
	elapsed := time.Since(start).Round(time.Millisecond)

	if err != nil {
		log.Error("", "exit", "failure", "error", err, "elapsed", elapsed)
		return 1
	}
	log.Info("", "exit", "success", "elapsed", elapsed)
	return 0
}

// root is a hack to ease testing.
func customizeMotd(log hclog.Logger, op string, rootDir string) error {
	now := time.Now().Round(time.Second)
	line := fmt.Sprintf("%s System %s by 🌼 florist 🌺\n", now, op)
	name := path.Join(rootDir, "/etc/motd")
	log.Debug("customizeMotd", "target", name, "operation", op)

	if err := florist.Mkdir(path.Dir(name), florist.User().Username, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	_, errWrite := f.WriteString(line)
	errClose := f.Close()
	return multierr.New(errWrite, errClose)
}
