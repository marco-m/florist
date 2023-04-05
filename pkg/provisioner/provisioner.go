// Package installer provides helper to write your own flower installer.
package provisioner

import (
	"fmt"
	"io/fs"
	stdlog "log"
	"sort"
	"time"

	"github.com/alecthomas/kong"
	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/pkg/florist"
)

type Provisioner struct {
	log           hclog.Logger
	cacheValidity time.Duration
	bouquets      map[string]Bouquet
	files         fs.FS
	secrets       fs.FS
	root          string
}

type Bouquet struct {
	Name        string
	Description string
	Flowers     []florist.Flower
}

func New(log hclog.Logger, cacheValidity time.Duration, files fs.FS, secrets fs.FS,
) (*Provisioner, error) {
	florist.SetLogger(log)

	stdlog.SetOutput(log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	stdlog.SetPrefix("")
	stdlog.SetFlags(0)

	return &Provisioner{
		log:           log.Named("installer"),
		cacheValidity: cacheValidity,
		bouquets:      map[string]Bouquet{},
		files:         files,
		secrets:       secrets,
	}, nil
}

func (prov *Provisioner) UseWorkdir() {
	prov.root = florist.WorkDir
}

func (prov *Provisioner) Run() error {
	var cli cli
	ctx := kong.Parse(
		&cli,
		kong.Description("🌼 florist 🌺 - a simple provisioner"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
	)

	// Invoke the Run method of the command passed on the command-line
	// (see the [cli] type).
	return ctx.Run(prov)
}

// Bouquets returns a list of the added bouquets, sorted by name.
func (prov *Provisioner) Bouquets() []Bouquet {
	// sort flowers in lexical order
	sortedNames := make([]string, 0, len(prov.bouquets))
	for name := range prov.bouquets {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	bouquets := make([]Bouquet, 0, len(prov.bouquets))
	for _, name := range sortedNames {
		bouquets = append(bouquets, prov.bouquets[name])
	}
	return bouquets
}

// AddBouquet creates a bouquet with `name` and `description` and adds `flowers` to it.
func (prov *Provisioner) AddBouquet(
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

	if _, ok := prov.bouquets[name]; ok {
		return fmt.Errorf("AddBouquet: there is already a bouquet with name %s", name)
	}

	prov.bouquets[name] = Bouquet{
		Name:        name,
		Description: description,
		Flowers:     flowers,
	}

	return nil
}
