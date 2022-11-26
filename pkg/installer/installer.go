// Package installer provides helper to write your own flower installer.
package installer

import (
	"fmt"
	"io/fs"
	stdlog "log"
	"os"
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
	fs            fs.FS
}

type Bouquet struct {
	Name        string
	Description string
	Flowers     []florist.Flower
}

func New(log hclog.Logger, cacheValidity time.Duration, fs fs.FS) Installer {
	florist.SetLogger(log)

	stdlog.SetOutput(log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	stdlog.SetPrefix("")
	stdlog.SetFlags(0)

	return Installer{
		log:           log.Named("installer"),
		cacheValidity: cacheValidity,
		bouquets:      map[string]Bouquet{},
		fs:            fs,
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
		if err := fl.Init(); err != nil {
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

type cliArgs struct {
	Install   *InstallCmd   `arg:"subcommand:install" help:"install one or more bouquets"`
	List      *ListCmd      `arg:"subcommand:list" help:"list the available bouquets"`
	EmbedList *EmbedListCmd `arg:"subcommand:embed-list" help:"list the embedded FS"`
}

func (cliArgs) Description() string {
	return "🌼 florist 🌺 - a simple installer\n"
}

type InstallCmd struct {
	Flower        []string `arg:"required,positional" help:"list of bouquets to install"`
	IgnoreUnknown bool     `arg:"--ignore-unknown" help:"ignore unknown bouquets instead of failing"`
}

type ListCmd struct {
	// TODO Petals bool `help:"List also details (each petal)"`
}

type EmbedListCmd struct{}

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
	case cliArgs.EmbedList != nil:
		return inst.cmdEmbedList()
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
		}
	}

	inst.log.Info("Customize motd")
	motd := "System provisioned by 🌼 florist 🌺\n"
	return os.WriteFile("/etc/motd", []byte(motd), 0644)
}

func (inst *Installer) cmdEmbedList() error {
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

	return fs.WalkDir(inst.fs, ".", fn)
}
