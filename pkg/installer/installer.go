// Package installer provides helper to write your own flower installer.
package installer

import (
	"fmt"
	"io/fs"
	stdlog "log"
	"os"
	"sort"
	"time"

	"github.com/alecthomas/kong"
	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/apt"
)

type Installer struct {
	log           hclog.Logger
	cacheValidity time.Duration
	bouquets      map[string]Bouquet
	fsys          fs.FS
}

type Bouquet struct {
	Name        string
	Description string
	Flowers     []florist.Flower
}

func New(log hclog.Logger, cacheValidity time.Duration, fsys fs.FS) Installer {
	florist.SetLogger(log)

	stdlog.SetOutput(log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	stdlog.SetPrefix("")
	stdlog.SetFlags(0)

	return Installer{
		log:           log.Named("installer"),
		cacheValidity: cacheValidity,
		bouquets:      map[string]Bouquet{},
		fsys:          fsys,
	}
}

// AddFlower adds a bouquet made of a single `flower`.
// See also: [AddBouquet].
func (inst *Installer) AddFlower(flower florist.Flower) error {
	return inst.AddBouquet(flower.String(), flower.Description(), flower)
}

// AddBouquet creates a bouquet with `name` and `description` and adds `flowers` to it.
// See also: [AddFlower].
func (inst *Installer) AddBouquet(
	name string,
	description string,
	flowers ...florist.Flower,
) error {
	if name == "" {
		return fmt.Errorf("AddBouquet: name cannot be empty")
	}
	if description == "" {
		return fmt.Errorf("AddBouquet %s: description cannot be empty", name)
	}
	if len(flowers) == 0 {
		return fmt.Errorf("AddBouquet %s: bouquet cannot be empty", name)
	}
	for i, fl := range flowers {
		if fl.String() == "" {
			return fmt.Errorf("AddBouquet %s: flower at position %d has empty name",
				name, i)
		}
		if fl.Description() == "" {
			return fmt.Errorf("AddBouquet %s: flower %s has empty description", name, fl)
		}
		if err := fl.Init(inst.fsys); err != nil {
			return fmt.Errorf("AddBouquet %s: flower %s: %s", name, fl, err)
		}
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

type cli struct {
	Install   InstallCmd   `cmd:"" help:"install one or more bouquets"`
	Configure ConfigureCmd `cmd:"" help:"configure one or more bouquets"`
	List      ListCmd      `cmd:"" help:"list the available bouquets"`
	EmbedList EmbedListCmd `cmd:"" help:"list the embedded FS"`
}

type InstallCmd struct {
	Bouquets []string `arg:"" help:"list of bouquets to install"`
}

type ConfigureCmd struct {
	Bouquets []string `arg:"" help:"list of bouquets to configure"`
}

type ListCmd struct {
	// TODO Petals bool `help:"List also details (each petal)"`
}

type EmbedListCmd struct{}

func (inst *Installer) Run() error {
	var cli cli
	ctx := kong.Parse(
		&cli,
		kong.Description("🌼 florist 🌺 - a simple installer"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
	)

	return ctx.Run(inst)
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

func (cmd *ListCmd) Run(inst *Installer) error {
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

func (cmd *InstallCmd) Run(inst *Installer) error {
	for _, name := range cmd.Bouquets {
		if _, ok := inst.bouquets[name]; !ok {
			return fmt.Errorf("install: unknown bouquet %s", name)
		}
	}

	if _, err := florist.Init(); err != nil {
		return err
	}

	inst.log.Info("Update package cache")
	if err := apt.Update(inst.cacheValidity); err != nil {
		return err
	}

	for _, name := range cmd.Bouquets {
		bouquet := inst.bouquets[name]
		inst.log.Info("installing", "bouquet", name, "flowers", len(bouquet.Flowers))
		for _, flower := range bouquet.Flowers {
			inst.log.Info("Installing", "flower", flower.String())
			if err := flower.Install(); err != nil {
				return err
			}
		}
	}

	inst.log.Info("Customize motd")
	motd := "System installed by 🌼 florist 🌺\n"
	return os.WriteFile("/etc/motd", []byte(motd), 0644)
}

func (cmd *ConfigureCmd) Run(inst *Installer) error {
	for _, name := range cmd.Bouquets {
		if _, ok := inst.bouquets[name]; !ok {
			return fmt.Errorf("configure: unknown bouquet %s", name)
		}
	}

	if _, err := florist.Init(); err != nil {
		return err
	}

	for _, name := range cmd.Bouquets {
		bouquet := inst.bouquets[name]
		inst.log.Info("configuring", "bouquet", name, "flowers", len(bouquet.Flowers))
		for _, flower := range bouquet.Flowers {
			inst.log.Info("configuring", "flower", flower.String())
			if err := flower.Configure(); err != nil {
				return err
			}
		}
	}

	inst.log.Info("Customize motd")
	motd := "System configured by 🌼 florist 🌺\n"
	return os.WriteFile("/etc/motd", []byte(motd), 0644)
}

func (cmd *EmbedListCmd) Run(inst *Installer) error {
	fn := func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		kind := "f"
		if de.IsDir() {
			kind = "d"
		}
		fmt.Println(kind, path)
		return nil
	}

	if err := fs.WalkDir(inst.fsys, ".", fn); err != nil {
		return fmt.Errorf("embed-list: fsys: %s", err)
	}

	return nil
}
