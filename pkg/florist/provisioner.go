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
	"time"

	"github.com/marco-m/clim"
)

const (
	DefOsPkgCacheValidity = 1 * time.Hour
)

var (
	// currentUser is set by Main.
	currentUser *user.User
	// osPkgCacheValidity is set by Main.
	osPkgCacheValidity time.Duration
	// floristLog is the logger for the whole florist module. Set by Main.
	floristLog *slog.Logger
)

// The Options passed to [MainInt]. For an example, see florist/example/main.go
type Options struct {
	// Output for the logger. Defaults to os.Stdout. Before changing to os.Stderr,
	// consider that HashiCorp Packer renders any output to stderr in red, thus
	// making everything look like an error.
	// The default log level is INFO; it can be changed to DEBUG via the --log-level
	// command-line flag.
	LogOutput io.Writer
	// Optimization to avoid refreshing the OS package manager cache each time before
	// installing an OS package. Defaults to DefOsPkgCacheValidity.
	OsPkgCacheValidity time.Duration
	// Set to a temporary directory during testing. DO NOT MODIFY in production code.
	RootDir string
	// The setup function, called before any command-line subcommand. Mandatory.
	SetupFn func(prov *Provisioner) error
	// The preConfigure function, called before the command-line configure
	// subcommand. Mandatory.
	PreConfigureFn func(prov *Provisioner, config *Config) (any, error)
	// The postConfigure function, called after the command-line configure
	// subcommand. Optional.
	PostConfigureFn func(prov *Provisioner, config *Config, bag any) error
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
	err := MainErr(os.Args, opts)
	if err == nil {
		return 0
	}
	if errors.Is(err, clim.ErrHelp) {
		fmt.Fprintln(os.Stdout, err)
		return 0
	}
	fmt.Fprintln(os.Stderr, err)
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
// See also [MainInt].
func MainErr(args []string, opts *Options) error {
	prog := filepath.Base(os.Args[0])
	app := App{
		start: time.Now(),
		prov:  newProvisioner(),
		opts:  opts,
	}

	cli, err := clim.NewTop[App](prog, "A 🌼 florist 🌺 provisioner", nil)
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

	if err := newListCmd(cli); err != nil {
		return err
	}
	if err := newInstallCmd(cli); err != nil {
		return err
	}
	if err := newConfigureCmd(cli); err != nil {
		return err
	}

	action, err := cli.Parse(args[1:])
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
	if opts.OsPkgCacheValidity == 0 {
		opts.OsPkgCacheValidity = DefOsPkgCacheValidity
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

	if err := LowLevelInit(opts.LogOutput, app.LogLevel, opts.OsPkgCacheValidity); err != nil {
		return err
	}

	app.log = Log()

	if err := opts.SetupFn(app.prov); err != nil {
		return fmt.Errorf("florist.Main: setup: %s", err)
	}

	return action(app)
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

type Provisioner struct {
	flowers map[string]Flower
	ordered []string
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
		return fmt.Errorf("Provisioner.AddFlowers: cannot call more than once")
	}
	for i, flower := range flowers {
		if flower.String() == "" {
			return fmt.Errorf("Provisioner.AddFlowers: flower %d has empty name", i)
		}
		if flower.Description() == "" {
			return fmt.Errorf("Provisioner.AddFlowers: flower %s has empty description",
				flower)
		}
		if _, found := prov.flowers[flower.String()]; found {
			return fmt.Errorf("Provisioner.AddFlowers: flower with same name already exists: %s", flower)
		}
		prov.ordered = append(prov.ordered, flower.String())
		prov.flowers[flower.String()] = flower
	}
	return nil
}

// root is a hack to ease testing.
func customizeMotd(op string, rootDir string) error {
	log := Log()
	now := time.Now().Round(time.Second)
	line := fmt.Sprintf("%s 🌼 florist 🌺 System %s\n", now, op)
	name := path.Join(rootDir, "/etc/motd")
	log.Debug("customize-motd", "target", name, "operation", op)

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

func CacheValidity() time.Duration {
	if osPkgCacheValidity == 0 {
		panic("florist.CacheValidity: must call florist.MainInt before")
	}
	return osPkgCacheValidity
}

func Log() *slog.Logger {
	if floristLog == nil {
		panic("florist.Log: must call florist.MainInt before")
	}
	return floristLog
}

// LowLevelInit should be called only by low-level test code.
// Absolutely do not call in non-test code! Call florist.MainInt instead!
func LowLevelInit(logOutput io.Writer, logLevel string, cacheValidity time.Duration) error {
	var level slog.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		return fmt.Errorf("florist.Main: --log-level: %s", err)
	}

	prog := filepath.Base(os.Args[0])
	floristLog = slog.New(slog.NewTextHandler(logOutput,
		&slog.HandlerOptions{Level: level})).With("prog", prog)

	// FIXME should this go below???
	var err error
	currentUser, err = user.Current()
	if err != nil {
		return fmt.Errorf("florist.Main: %s", err)
	}

	osPkgCacheValidity = cacheValidity

	return nil
}
