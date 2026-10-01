package florist

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/marco-m/clim"
	"github.com/marco-m/florist/internal"
)

var (
	// currentUser is set by [MainInt].
	currentUser *user.User
)

// The Options passed to [MainInt]. For an example, see florist/example/main.go
type Options struct {
	// Seeds will be passed to each [Flower.Install] and [Flower.Configure].
	Seeds

	// LogOutput is the output of the logger. Defaults to [os.Stdout].
	// Before changing to [os.Stderr], consider that HashiCorp Packer renders
	// any output to stderr in red, thus  making everything look like an error.
	LogOutput io.Writer
	// SetupFn will be called before any command-line subcommand. Mandatory.
	SetupFn func(prov *Provisioner) error
	// PreConfigureFn will be called before the command-line "configure".
	// Mandatory.
	PreConfigureFn func(prov *Provisioner, config *Config) (any, error)
	// PostConfigureFn will be called after the command-line "configure".
	// Optional.
	PostConfigureFn func(prov *Provisioner, config *Config, bag any) error
}

// Seeds will be passed to each [Flower.Init] and [Flower.Configure] by [MainInt].
type Seeds struct {
	// RootDir can be set to a temporary directory during testing.
	// DO NOT MODIFY in production code.
	RootDir string
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

// ExitCode can be used with [MainErr]. See also [MainInt].
func ExitCode(err error) int {
	if err == nil || errors.Is(err, clim.ErrHelp) {
		return 0
	}
	return 1
}

type App struct {
	LogLevel string
	//
	start time.Time
	log   *slog.Logger
	prov  *Provisioner
	opts  *Options
}

// MainErr is a ready-made function for the main() of your installer.
// See also [MainInt] and [ExitCode].
func MainErr(args []string, opts *Options) error {
	prog := filepath.Base(os.Args[0])
	app := App{
		start: time.Now(),
		prov:  newProvisioner(),
		opts:  opts,
	}

	cli, err := clim.NewTop(prog, "A 🌼 florist 🌺 provisioner")
	if err != nil {
		return err
	}

	if err := cli.AddFlags(
		&clim.Flag{
			Value: clim.String(&app.LogLevel, "INFO"),
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

	if opts.LogOutput == nil {
		opts.LogOutput = os.Stdout
	}
	if opts.RootDir == "" {
		opts.RootDir = "/"
	}
	if opts.SetupFn == nil {
		return fmt.Errorf("florist.Main: SetupFn is nil")
	}
	if opts.PreConfigureFn == nil {
		return fmt.Errorf("florist.Main: PreConfigureFn is nil")
	}

	if err := LowLevelInit(opts.LogOutput, app.LogLevel); err != nil {
		return err
	}

	app.log = slog.Default()

	if err := opts.SetupFn(app.prov); err != nil {
		return fmt.Errorf("florist.Main: setup: %s", err)
	}

	_, subcommand, _ := strings.CutLast(command, " ")
	switch subcommand {
	case "list":
		return listCmd.Run(app)
	case "install":
		return installCmd.Run(app)
	case "configure":
		return configureCmd.Run(app)
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

// Flowers returns
func (prov *Provisioner) Flowers() map[string]Flower {
	return prov.flowers
}

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

// User returns the current user, as set by Init.
func User() *user.User {
	if currentUser == nil {
		panic("florist.User: must call florist.MainInt before")
	}
	return currentUser
}

// Group returns the primary group of the current user, as set by Init.
func Group() *user.Group {
	if currentUser == nil {
		panic("florist.Group: must call florist.MainInt before")
	}
	group, _ := user.LookupGroupId(currentUser.Gid)
	return group
}

// LowLevelInit should be called only by low-level test code.
// Absolutely do not call in non-test code! Call florist.MainInt instead!
func LowLevelInit(logOutput io.Writer, logLevel string) error {
	errorf := internal.MakeErrorf("florist.LowLevelInit")
	var level slog.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		return errorf("--log-level: %s", err)
	}

	prog := filepath.Base(os.Args[0])
	slog.SetDefault(slog.New(slog.NewTextHandler(logOutput,
		&slog.HandlerOptions{Level: level})).With("prog", prog))

	// FIXME should this go below???
	var err error
	currentUser, err = user.Current()
	if err != nil {
		return errorf("%s", err)
	}

	if err := Mkdir(WorkDir, 0o755, User().Username, Group().Name); err != nil {
		return errorf("%s", err)
	}
	// if err := Mkdir(HomeDir, 0o755, User().Username, Group().Name); err != nil {
	// 	return errorf("%s", err)
	// }

	return nil
}

// root is a hack to ease testing.
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

func timelog(run func() error, app App) error {
	app.log.Info("starting", "command-line", os.Args)
	err := run()
	elapsed := time.Since(app.start).Round(time.Millisecond)
	if err != nil {
		app.log.Error("exiting", "status", "failure", "error", err, "elapsed", elapsed)
		return err
	}
	app.log.Info("exiting", "status", "success", "elapsed", elapsed)
	return nil
}
