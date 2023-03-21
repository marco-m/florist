package installer

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist"
	"io/fs"
	"os"
	"path"
	"strings"
	"tailscale.com/util/multierr"
	"time"
)

type cli struct {
	Install   InstallCmd   `cmd:"" help:"install a bouquet"`
	Configure ConfigureCmd `cmd:"" help:"configure a bouquet"`
	List      ListCmd      `cmd:"" help:"list the available bouquets"`
	EmbedList EmbedListCmd `cmd:"" help:"list the embedded FS"`
}

type InstallCmd struct {
	Bouquet string `arg:"" help:"Bouquet to install"`
}

func (cmd *InstallCmd) Run(inst *Installer) error {
	if _, err := florist.Init(); err != nil {
		return err
	}

	bouquet, ok := inst.bouquets[cmd.Bouquet]
	if !ok {
		return fmt.Errorf("install: unknown bouquet %s", cmd.Bouquet)
	}

	inst.log.Info("installing", "bouquet", cmd.Bouquet,
		"size", len(bouquet.Flowers), "flowers", bouquet.Flowers)

	for _, flower := range bouquet.Flowers {
		inst.log.Debug("before", "files", describeFS(inst.files))
		files2, err := fs.Sub(inst.files, flower.String())
		if err != nil {
			return fmt.Errorf("install: %s", err)
		}
		secrets2, err := fs.Sub(inst.secrets, "flowers/"+flower.String())
		if err != nil {
			return fmt.Errorf("configure: %s", err)
		}
		finder := florist.NewFsFinder(secrets2)

		inst.log.Info("Installing", "flower", flower.String(),
			"files", describeFS(files2), "secrets", describeFS(secrets2))
		if err := flower.Install(files2, finder); err != nil {
			return err
		}
	}

	return customizeMotd(inst.log, "installed", inst.root)
}

type ConfigureCmd struct {
	Bouquet string `arg:"" help:"Bouquet to configure (assumes that the same bouquet is installed on the image"`
}

func (cmd *ConfigureCmd) Run(inst *Installer) error {
	if _, err := florist.Init(); err != nil {
		return err
	}

	bouquet, ok := inst.bouquets[cmd.Bouquet]
	if !ok {
		return fmt.Errorf("configure: unknown bouquet %s", cmd.Bouquet)
	}

	inst.log.Info("configuring", "bouquet", cmd.Bouquet,
		"size", len(bouquet.Flowers), "flowers", bouquet.Flowers)

	for _, flower := range bouquet.Flowers {
		files2, err := fs.Sub(inst.files, flower.String())
		if err != nil {
			return fmt.Errorf("configure: %s", err)
		}

		// secrets can be in two places:
		// - flowers/FLOWER/k1
		// - nodes/NODE/FLOWER/k2
		// where the node overrides the flower
		secretsA, err := fs.Sub(inst.secrets, path.Join("nodes", cmd.Bouquet, flower.String()))
		if err != nil {
			return fmt.Errorf("configure: %s", err)
		}
		finderA := florist.NewFsFinder(secretsA)

		secretsB, err := fs.Sub(inst.secrets, path.Join("flowers", flower.String()))
		if err != nil {
			return fmt.Errorf("configure: %s", err)
		}
		finderB := florist.NewFsFinder(secretsB)

		finder := florist.NewUnionFinder(finderA, finderB)
		inst.log.Info("configuring", "flower", flower.String(),
			"files", describeFS(files2), "secretsA", describeFS(secretsA),
			"secretsB", describeFS(secretsB))
		if err := flower.Configure(files2, finder); err != nil {
			return err
		}
	}

	return customizeMotd(inst.log, "configured", inst.root)
}

type ListCmd struct {
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

type EmbedListCmd struct{}

func (cmd *EmbedListCmd) Run(inst *Installer) error {
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

	fmt.Fprintln(&bld, "files:")
	if err := fs.WalkDir(files, ".", fn); err != nil {
		return bld.String(), fmt.Errorf("embed-list: files: %s", err)
	}
	fmt.Fprintln(&bld, "secrets:")
	if err := fs.WalkDir(secrets, ".", fn); err != nil {
		return bld.String(), fmt.Errorf("embed-list: secrets: %s", err)
	}

	return bld.String(), nil
}

// PrepareFn is the function signature to pass to [Main].
type PrepareFn func(log hclog.Logger) (*Installer, error)

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
		return inst.Run()
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
