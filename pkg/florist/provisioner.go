package florist

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/marco-m/clim"
)

// The Options passed to [MainInt]. For an example, see florist/example/main.go
type Options struct {
	// The SetupFn user callback will be invoked before any command-line
	// subcommand. It is meant to call [Provisioner.AddFlowers].
	// Mandatory.
	SetupFn func(prov *Provisioner) error
	// The PreConfigureFn user callback will be invoked before the command-line
	// "configure". It is meant to configure the flowers that have been added
	// by [Options.SetupFn], referencing the K/V pairs of [Config].
	// Mandatory.
	PreConfigureFn func(prov *Provisioner, config *Config) (any, error)
	// The PostConfigureFn user callback will be invoked after the command-line
	// "configure". Optional.
	PostConfigureFn func(prov *Provisioner, config *Config, bag any) error

	// LogOutput is the output of the logger. Defaults to [os.Stdout].
	// Before changing to [os.Stderr], consider that HashiCorp Packer renders
	// any output to stderr in red, thus  making everything look like an error.
	LogOutput io.Writer

	// TempDir is the base temporary directory where to store downloaded
	// packages and similar. Defaults to [os.TempDir].
	// It is used to set [App.WorkDir] as filepath.Join(TempDir, "florist").
	TempDir string

	// RootDir can be set to a temporary directory during testing.
	// The default is /, the real root of the filesystems.
	// DO NOT MODIFY in production code.
	// WARNING: almost no code respects this parameter.
	RootDir string

	//
	LogLevel slog.Level
}

// MainInt is a ready-made function for the main() of your installer.
//
// Usage:
//
//	func main() {
//	    os.Exit(florist.MainInt(&florist.Options{
//	        SetupFn:         setup,
//	        PreConfigureFn:  preConfigure,
//	        PostConfigureFn: postConfigure
//	    }))
//	}
//
// See also [MainErr].
func MainInt(opts *Options) int {
	if err := MainErr(os.Args, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ExitCode(err)
	}
	return 0
}

// ExitCode can be used with [MainErr].
// If using [MainInt], you do not need this function.
func ExitCode(err error) int {
	if err == nil || errors.Is(err, clim.ErrHelp) {
		return 0
	}
	return 1
}

// MainErr is a ready-made function for the main() of your installer.
// See also [MainInt] and [ExitCode].
func MainErr(args []string, opts *Options) error {
	prog := filepath.Base(os.Args[0])
	cli, err := clim.NewTop(prog, "A 🌼 florist 🌺 provisioner")
	if err != nil {
		return err
	}

	if err := cli.AddFlags(
		&clim.Flag{
			Value: clim.LogLevel(&opts.LogLevel, slog.LevelInfo),
			Long:  "log-level", Help: "set the log level",
		}); err != nil {
		return err
	}

	listCmd, err := newListCmd(cli)
	if err != nil {
		return err
	}
	installCmd, err := newInstallCmd(cli)
	if err != nil {
		return err
	}
	configureCmd, err := newConfigureCmd(cli)
	if err != nil {
		return err
	}

	command, err := cli.Parse(args[1:])
	if err != nil {
		return err
	}

	// FIXME some tests trip on this. I should re-enable the check and use an explicit
	// TestMain ?
	// if currentUser != nil {
	// 	return fmt.Errorf("florist.Main: function already called")
	// }

	if opts.SetupFn == nil {
		return fmt.Errorf("florist.Main: SetupFn is nil")
	}
	if opts.PreConfigureFn == nil {
		return fmt.Errorf("florist.Main: PreConfigureFn is nil")
	}

	gdn, err := NewGarden(opts)
	if err != nil {
		return err
	}

	if err := opts.SetupFn(gdn.prov); err != nil {
		return fmt.Errorf("florist.Main: setup: %s", err)
	}

	_, subcommand, _ := strings.CutLast(command, " ")
	switch subcommand {
	case "list":
		return listCmd.Run(gdn)
	case "install":
		return installCmd.Run(gdn)
	case "configure":
		return configureCmd.Run(gdn)
	default:
		return fmt.Errorf("internal error: unwired command: %s", command)
	}
}

type Provisioner struct {
	flowers map[string]Flower
	ordered []string
	errs    []error
}

func (prov *Provisioner) Errors() []error {
	dst := make([]error, len(prov.errs))
	copy(dst, prov.errs)
	return dst
}

func newProvisioner() *Provisioner {
	return &Provisioner{
		flowers: make(map[string]Flower),
	}
}

// Flowers returns the flowers known to the provisioner.
// See also: [Provisioner.AddFlowers].
func (prov *Provisioner) Flowers() map[string]Flower {
	return prov.flowers
}

// AddFlowers adds flowers to the provisioner.
// Meant to be called from the [Options.SetupFn] callback of the user program.
func (prov *Provisioner) AddFlowers(flowers ...Flower) error {
	if len(prov.flowers) > 0 {
		return fmt.Errorf("florist.AddFlowers: cannot call more than once")
	}
	for i, flower := range flowers {
		if flower.String() == "" {
			return fmt.Errorf("florist.AddFlowers: flower %d has empty name", i)
		}
		if flower.Description() == "" {
			return fmt.Errorf("florist.AddFlowers: flower %s has empty description",
				flower)
		}
		if _, found := prov.flowers[flower.String()]; found {
			return fmt.Errorf("florist.AddFlowers: flower with same name already exists: %s", flower)
		}
		prov.ordered = append(prov.ordered, flower.String())
		prov.flowers[flower.String()] = flower
	}
	return nil
}

// rootDir is a hack to ease testing. See [Options.RootDir].
func customizeMotd(op string, status string, rootDir string) error {
	now := time.Now().UTC().Round(time.Second)
	line := fmt.Sprintf("%s 🌼 florist 🌺 System %s (%s)\n", now, op, status)
	name := path.Join(rootDir, "/etc/motd")
	slog.Debug("customize-motd", "target", name, "operation", op, "status", status)

	if err := os.MkdirAll(path.Dir(name), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	_, errWrite := f.WriteString(line)
	errClose := f.Close()
	return JoinErrors(errWrite, errClose)
}

func timelog(run func() error, gdn *Garden) error {
	gdn.log.Info("starting", "command-line", os.Args)
	err := run()
	elapsed := time.Since(gdn.start).Round(time.Millisecond)
	if err != nil {
		gdn.log.Error("exiting", "status", "failure", "error", err, "elapsed", elapsed)
		return err
	}
	gdn.log.Info("exiting", "status", "success", "elapsed", elapsed)
	return nil
}
