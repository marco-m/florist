// Package installer provides helper to write your own flower installer.
package installer

import (
	"fmt"
	stdlog "log"
	"sort"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
)

type Installer struct {
	flowers map[string]florist.Flower
	log     hclog.Logger
	// fixme this should be abstracted out somehow...
	aptCacheValidity time.Duration
}

func New(log hclog.Logger, cacheValidity time.Duration) *Installer {
	florist.SetLogger(log)

	stdlog.SetOutput(log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	stdlog.SetPrefix("")
	stdlog.SetFlags(0)

	return &Installer{
		flowers:          map[string]florist.Flower{},
		log:              log.Named("installer"),
		aptCacheValidity: cacheValidity,
	}
}

func (inst *Installer) AddFlower(name string, flower florist.Flower) error {
	if _, ok := inst.flowers[name]; ok {
		return fmt.Errorf("a flower with name %s exists already", name)
	}
	flower.SetLogger(florist.Log)
	inst.flowers[name] = flower
	return nil
}

type cliArgs struct {
	Install *InstallCmd `arg:"subcommand:install" help:"install one or more flowers in sequence"`
	List    *ListCmd    `arg:"subcommand:list" help:"list the available flowers"`
}

func (cliArgs) Description() string {
	return "🌼 florist 🌺 - a simple installer\n"
}

type InstallCmd struct {
	Flower        []string `arg:"required,positional" help:"list of flowers to install"`
	IgnoreUnknown bool     `arg:"--ignore-unknown" help:"ignore unknown flowers instead of failing"`
}

type ListCmd struct { //
	// TODO Petals bool `help:"List also details (each petal)"`
}

func (inst *Installer) Run() error {
	var cliArgs cliArgs

	parser := arg.MustParse(&cliArgs)
	if parser.Subcommand() == nil {
		parser.Fail("missing subcommand")
	}

	switch {
	case cliArgs.Install != nil:
		return inst.cmdInstall(cliArgs.Install.Flower, cliArgs.Install.IgnoreUnknown)
	case cliArgs.List != nil:
		return inst.cmdList()
	default:
		return fmt.Errorf("internal error: unwired subcommand: %s", parser.SubcommandNames()[0])
	}
}

func (inst *Installer) cmdList() error {
	// sort keys in lexical order
	keys := make([]string, 0, len(inst.flowers))
	for k := range inst.flowers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Println("Available flowers:")
	for _, name := range keys {
		fmt.Printf("%-15s %s\n", name, inst.flowers[name])
	}
	return nil
}

func (inst *Installer) cmdInstall(names []string, ignore bool) error {
	existing := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := inst.flowers[name]; !ok {
			if ignore {
				inst.log.Warn("ignoring unknown", "flower", name)
				continue
			}
			return fmt.Errorf("install: unknown flower %s", name)
		}
		existing = append(existing, name)
	}

	if len(existing) == 0 {
		inst.log.Warn("all flowers are unknown, nothing to do")
		return nil
	}

	if _, err := florist.Init(); err != nil {
		return err
	}

	inst.log.Info("Update package cache")
	if err := apt.Update(inst.aptCacheValidity); err != nil {
		return err
	}

	for _, name := range existing {
		flower := inst.flowers[name]
		inst.log.Info("Install", "flower", name)

		if err := flower.Install(); err != nil {
			return err
		}

		if err := florist.WriteRecord(name); err != nil {
			inst.log.Warn("WriteRecord", "error", err)
		}
	}
	return nil
}
