// Package provisioner provides helper to write your own florist provisioner.
package provisioner

import (
	"errors"
	"fmt"
	"io/fs"
	stdlog "log"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/hashicorp/go-hclog"
	"tailscale.com/util/multierr"

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

	if files == nil {
		return nil, errors.New("provisioner.New: nil files")
	}
	if secrets == nil {
		return nil, errors.New("provisioner.New: nil secrets")
	}

	return &Provisioner{
		log:           log.Named("provisioner"),
		cacheValidity: cacheValidity,
		bouquets:      map[string]Bouquet{},
		files:         files,
		secrets:       secrets,
	}, nil
}

func (prov *Provisioner) UseWorkdir() {
	prov.root = florist.WorkDir
}

func (prov *Provisioner) Run(args []string) error {
	var cli cli
	parser, err := kong.New(&cli,
		kong.Description("🌼 florist 🌺 - a simple provisioner"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
	)
	if err != nil {
		panic(err)
	}
	ctx, err := parser.Parse(args)
	parser.FatalIfErrorf(err)

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

type cli struct {
	Install   InstallCmd   `cmd:"" help:"install a bouquet"`
	Configure ConfigureCmd `cmd:"" help:"configure a bouquet"`
	List      ListCmd      `cmd:"" help:"list the available bouquets"`
	EmbedList EmbedListCmd `cmd:"" help:"list the embedded FS"`
}

type InstallCmd struct {
	Bouquet string `arg:"" help:"Bouquet to install"`
}

func (cmd *InstallCmd) Run(prov *Provisioner) error {
	if _, err := florist.Init(); err != nil {
		return err
	}

	bouquet, ok := prov.bouquets[cmd.Bouquet]
	if !ok {
		return fmt.Errorf("install: unknown bouquet %s", cmd.Bouquet)
	}

	prov.log.Info("installing", "bouquet", cmd.Bouquet,
		"size", len(bouquet.Flowers), "flowers", bouquet.Flowers)

	for _, flower := range bouquet.Flowers {
		prov.log.Info("Installing", "flower", flower.String())
		filesPerFlower, _, secretsPerFlower, _, err :=
			subFS(prov.log, flower.String(), cmd.Bouquet, prov.files, prov.secrets)
		if err != nil {
			return fmt.Errorf("install: %s", err)
		}
		finderFlowers := florist.NewFsFinder(secretsPerFlower)

		if err := flower.Install(filesPerFlower, finderFlowers); err != nil {
			return err
		}
	}

	return customizeMotd(prov.log, "installed", prov.root)
}

type ConfigureCmd struct {
	Bouquet string `arg:"" help:"Bouquet to configure (assumes that the same bouquet is installed on the image"`
}

func (cmd *ConfigureCmd) Run(prov *Provisioner) error {
	if _, err := florist.Init(); err != nil {
		return err
	}

	bouquet, ok := prov.bouquets[cmd.Bouquet]
	if !ok {
		return fmt.Errorf("configure: unknown bouquet %s", cmd.Bouquet)
	}

	prov.log.Info("configuring", "bouquet", cmd.Bouquet,
		"size", len(bouquet.Flowers), "flowers", bouquet.Flowers)

	for _, flower := range bouquet.Flowers {
		prov.log.Info("configuring", "flower", flower.String())
		filesPerFlower, secretsBase, secretsPerFlower, secretsPerNode, err :=
			subFS(prov.log, flower.String(), cmd.Bouquet, prov.files, prov.secrets)
		if err != nil {
			return fmt.Errorf("configure: %s", err)
		}

		// secrets can be in three places:
		// flowers/FLOWER/k1
		// nodes/NODE/FLOWER/k2
		// base/
		// where node overrides flower, flower overrides base.

		finderNodes := florist.NewFsFinder(secretsPerNode)
		finderFlowers := florist.NewFsFinder(secretsPerFlower)
		finderBase := florist.NewFsFinder(secretsBase)
		finder := florist.NewUnionFinder(finderNodes, finderFlowers, finderBase)
		if err := flower.Configure(filesPerFlower, finder); err != nil {
			return err
		}
	}

	return customizeMotd(prov.log, "configured", prov.root)
}

func subFS(log hclog.Logger, flower string, node string, files, secrets fs.FS,
) (
	filesPerFlower, secretsBase, secretsPerFlower, secretsPerNode fs.FS, err error,
) {
	log.Debug("subFS: contents before", "flower", flower, "node", node,
		"allFiles", describeFS(files),
		"allSecrets", describeFS(secrets))

	subdir := flower
	log.Debug("filesPerFlower", "subdir", subdir)
	filesPerFlower, err = fs.Sub(files, subdir)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	subdir = "base"
	log.Debug("secretsBase", "subdir", subdir)
	secretsBase, err = fs.Sub(secrets, subdir)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	subdir = path.Join("flowers", flower)
	log.Debug("secretsPerFlower", "subdir", subdir)
	secretsPerFlower, err = fs.Sub(secrets, subdir)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	subdir = path.Join("nodes", node, flower)
	log.Debug("secretsPerNode", "subdir", subdir)
	secretsPerNode, err = fs.Sub(secrets, subdir)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	log.Debug("subFS: contents after", "flower", flower, "node", node,
		"base", describeFS(secretsBase),
		"filesPerFlower", describeFS(filesPerFlower),
		"secretsPerFlower", describeFS(secretsPerFlower),
		"secretsPerNode", describeFS(secretsPerNode))

	return filesPerFlower, secretsBase, secretsPerFlower, secretsPerNode, nil
}

type ListCmd struct {
}

func (cmd *ListCmd) Run(inst *Provisioner) error {
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

type EmbedListCmd struct{}

func (cmd *EmbedListCmd) Run(inst *Provisioner) error {
	list, err := ListFs(inst.files, inst.secrets)
	fmt.Println(list)
	return err
}

func ListFs(files fs.FS, secrets fs.FS) (string, error) {
	var bld strings.Builder

	fn := func(dePath string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !de.IsDir() {
			fmt.Fprintln(&bld, dePath)
		}
		return nil
	}

	// Ignore errors.
	fmt.Fprintln(&bld, "files:")
	fs.WalkDir(files, ".", fn)
	fmt.Fprintln(&bld, "secrets:")
	fs.WalkDir(secrets, ".", fn)

	return bld.String(), nil
}

// PrepareFn is the function signature to pass to [Main].
type PrepareFn func(log hclog.Logger) (*Provisioner, error)

// Main is a ready-made function for the main() of your installer.
// Usage:
//
//	func main() {
//		log := florist.NewLogger("example")
//		os.Exit(installer.Main(log, prepare))
//	}
//
// If you don't want to use it, just copy it and modify as you see fit.
func Main(log hclog.Logger, prepare PrepareFn) int {
	log.Info("starting", "invocation", os.Args)
	start := time.Now()
	steps := func() error {
		inst, err := prepare(log)
		if err != nil {
			return err
		}
		return inst.Run(os.Args[1:])
	}
	err := steps()
	elapsed := time.Since(start).Round(time.Millisecond)

	if err != nil {
		log.Error("", "exit", "failure", "error", err, "elapsed", elapsed)
		return 1
	}
	log.Info("", "exit", "success", "elapsed", elapsed)
	return 0
}

// root is a hack to ease testing.
func customizeMotd(log hclog.Logger, op string, root string) error {
	now := time.Now().Round(time.Second)
	line := fmt.Sprintf("%s System %s by 🌼 florist 🌺\n", now, op)
	name := path.Join(root, "/etc/motd")
	log.Debug("customizeMotd", "target", name, "operation", op)

	if err := florist.Mkdir(path.Dir(name), nil, 0755); err != nil {
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

func describeFS(fsys fs.FS) string {
	matches, _ := fs.Glob(fsys, "*")
	return strings.Join(matches, ",")
}
