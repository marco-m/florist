// Package installer provides helper to write your own flower installer.
package installer

import (
	"errors"
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
	log           hclog.Logger
	cacheValidity time.Duration
	bouquets      map[string]Bouquet
}

type Bouquet struct {
	Name        string
	Description string
	Flowers     []florist.Flower
}

func New(log hclog.Logger, cacheValidity time.Duration) Installer {
	florist.SetLogger(log)

	stdlog.SetOutput(log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	stdlog.SetPrefix("")
	stdlog.SetFlags(0)

	return Installer{
		log:           log.Named("installer"),
		cacheValidity: cacheValidity,
		bouquets:      map[string]Bouquet{},
	}
}

// AddBouquet adds `flowers` with `name` and `description`.
func (inst *Installer) AddBouquet(
	name string,
	description string,
	flowers []florist.Flower,
) error {
	if len(flowers) == 0 {
		return errors.New("AddBouquet: bouquet is empty")
	}
	if len(flowers) > 1 {
		if name == "" {
			return fmt.Errorf(
				"AddBouquet: more that one flower and name is empty: %s", flowers)
		}
		if description == "" {
			return fmt.Errorf(
				"AddBouquet: more that one flower and description is empty: %s", flowers)
		}
	}
	for i, fl := range flowers {
		if fl.String() == "" {
			return fmt.Errorf("AddBouquet: flower %d has empty name", i)
		}
		if fl.Description() == "" {
			return fmt.Errorf("AddBouquet: flower %d has empty description", i)
		}
	}

	if name == "" {
		name = flowers[0].String()
	}
	if description == "" {
		description = flowers[0].Description()
	}

	if _, ok := inst.bouquets[name]; ok {
		return fmt.Errorf("AddBouquet: there is already a bouquet with name %s", name)
	}

	inst.bouquets[name] = Bouquet{
		Name:        name,
		Description: description,
		Flowers:     flowers,
	}

	return nil
}

type cliArgs struct {
	Install *InstallCmd `arg:"subcommand:install" help:"install one or more bouquets"`
	List    *ListCmd    `arg:"subcommand:list" help:"list the available bouquets"`
}

func (cliArgs) Description() string {
	return "🌼 florist 🌺 - a simple installer\n"
}

type InstallCmd struct {
	Flower        []string `arg:"required,positional" help:"list of bouquets to install"`
	IgnoreUnknown bool     `arg:"--ignore-unknown" help:"ignore unknown bouquets instead of failing"`
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

// Bouquets returns a list of the added bouquets sorted by name.
func (inst *Installer) Bouquets() []Bouquet {
	// sort flowers in lexical order
	sortedNames := make([]string, 0, len(inst.bouquets))
	for name := range inst.bouquets {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	bouquets := make([]Bouquet, 0, len(inst.bouquets))
	for _, name := range sortedNames {
		bouquets = append(bouquets, inst.bouquets[name])
	}
	return bouquets
}

func (inst *Installer) cmdList() error {
	fmt.Println("Available bouquets:")
	for _, bouquet := range inst.Bouquets() {
		fmt.Printf("%-20s %s\n", bouquet.Name, bouquet.Description)
		for _, fl := range bouquet.Flowers {
			fmt.Printf("    %-20s (%s)\n", fl.String(), fl.Description())
		}
		fmt.Println()
	}
	return nil
}

func (inst *Installer) cmdInstall(names []string, ignore bool) error {
	found := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := inst.bouquets[name]; !ok {
			if ignore {
				inst.log.Warn("ignoring unknown", "bouquet", name)
				continue
			}
			return fmt.Errorf("install: unknown bouquet %s", name)
		}
		found = append(found, name)
	}

	if len(found) == 0 {
		inst.log.Warn("all bouquets are unknown, nothing to do")
		return nil
	}

	if _, err := florist.Init(); err != nil {
		return err
	}

	inst.log.Info("Update package cache")
	if err := apt.Update(inst.cacheValidity); err != nil {
		return err
	}

	for _, name := range found {
		bouquet := inst.bouquets[name]
		inst.log.Info("installing", "bouquet", name, "flowers", len(bouquet.Flowers))
		for _, flower := range bouquet.Flowers {
			inst.log.Info("Installing", "flower", flower.String())
			if err := flower.Install(); err != nil {
				return err
			}
			if err := florist.WriteRecord(flower.String()); err != nil {
				inst.log.Warn("WriteRecord", "error", err)
			}
		}
	}
	return nil
}
