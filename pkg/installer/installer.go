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

type bouquet struct {
	flower     *florist.Flower
	subFlowers []*florist.Flower
}

type Installer struct {
	bouquets map[string]bouquet
	log      hclog.Logger
	// fixme this should be abstracted out somehow...
	aptCacheValidity time.Duration
}

func New(log hclog.Logger, cacheValidity time.Duration) *Installer {
	florist.SetLogger(log)

	stdlog.SetOutput(log.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	stdlog.SetPrefix("")
	stdlog.SetFlags(0)

	return &Installer{
		bouquets:         map[string]bouquet{},
		log:              log.Named("installer"),
		aptCacheValidity: cacheValidity,
	}
}

func (inst *Installer) newBouquet(
	flower *florist.Flower,
	subFlowers []string,
) (bouquet, error) {
	newBo := bouquet{
		flower:     flower,
		subFlowers: []*florist.Flower{},
	}
	for _, sb := range subFlowers {
		if bo, ok := inst.bouquets[sb]; !ok {
			return bouquet{}, fmt.Errorf(
				"install: unknown subflower %s (you must add it before", sb)
		} else {
			newBo.subFlowers = append(newBo.subFlowers, bo.flower)
		}
	}
	return newBo, nil
}

// AddFlower adds flower `flower` under name `name`.
// If subFlowers is not empty, it will add them as part of this flower.
// Note that subFlowers must have previously been added via AddFlower.
// FIXME it should accept pointers to subflowers instead of strings...
// FIXME actually I could pass the subflowers directly to the flower struct, seems simpler?
func (inst *Installer) AddFlower(
	flower florist.Flower,
	subFlowers ...string,
) error {
	name := flower.Description().Name
	if _, ok := inst.bouquets[name]; ok {
		return fmt.Errorf("a flower with name %s exists already", name)
	}
	flower.SetLogger(florist.Log)
	bo, err := inst.newBouquet(&flower, subFlowers)
	if err != nil {
		return err
	}
	inst.bouquets[name] = bo
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

// List returns a list of flowers, each flower with a list of subflowers.
func (inst *Installer) List() [][]string {
	// sort flowers in lexical order
	sortedNames := make([]string, 0, len(inst.bouquets))
	for k := range inst.bouquets {
		sortedNames = append(sortedNames, k)
	}
	sort.Strings(sortedNames)

	var list [][]string
	for _, name := range sortedNames {
		var elem []string
		bouquet := inst.bouquets[name]
		elem = append(elem, (*bouquet.flower).Description().Name)
		for _, subFl := range bouquet.subFlowers {
			elem = append(elem, (*subFl).Description().Name)
		}
		list = append(list, elem)
	}
	return list
}

func (inst *Installer) cmdList() error {
	fmt.Println("Available flowers:")
	for _, elem := range inst.List() {
		bouquet := inst.bouquets[elem[0]]
		flower := *bouquet.flower
		fmt.Printf("%-20s %s\n", flower.Description().Name, flower.Description().Long)
		for _, sb := range bouquet.subFlowers {
			fmt.Println("  ", (*sb).Description().Name)
		}
		fmt.Println()
	}
	return nil
}

func (inst *Installer) cmdInstall(names []string, ignore bool) error {
	existing := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := inst.bouquets[name]; !ok {
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
		bouquet := inst.bouquets[name]
		flower := *bouquet.flower
		inst.log.Info("Install", "main flower", name)
		if err := install(inst.log, flower); err != nil {
			return err
		}

		for _, subFlower := range bouquet.subFlowers {
			inst.log.Info("Install", "sub flower", (*subFlower).Description().Name)
			if err := install(inst.log, *subFlower); err != nil {
				return err
			}
		}

	}
	return nil
}

func install(log hclog.Logger, flower florist.Flower) error {
	if err := flower.Install(); err != nil {
		return err
	}
	if err := florist.WriteRecord(flower.Description().Name); err != nil {
		log.Warn("WriteRecord", "error", err)
	}
	return nil
}
