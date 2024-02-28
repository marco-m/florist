package florist

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"time"

	"github.com/alexflint/go-arg"
	"tailscale.com/util/multierr"
)

var (
	// currentUser is set by Main.
	currentUser *user.User
	// osPkgCacheValidity is set by Main.
	osPkgCacheValidity time.Duration
	// floristLog is the logger for the whole florist module. Set by Main.
	floristLog *slog.Logger
)

type cliArgs struct {
	LogLevel  string         `arg:"--log-level" help:"log level" placeholder:"LEVEL" default:"INFO"`
	List      *listArgs      `arg:"subcommand:list" help:"list the flowers and their files"`
	Install   *installArgs   `arg:"subcommand:install"`
	Configure *configureArgs `arg:"subcommand:configure"`
}

func (cliArgs) Description() string {
	prog := filepath.Base(os.Args[0])
	return fmt.Sprintf("%s -- A 🌼 florist 🌺 provisioner.", prog)
}

type ConfigureFn func(prov *Provisioner, config *Config) error

// The Options passed to Main. For an example, see florist/example/main.go
type Options struct {
	// Output for the logger. Defaults to os.Stdout. Before changing to os.Stderr,
	// consider that HashiCorp Packer renders any output to stderr in red, thus
	// making everything look like an error.
	// The default log level is INFO; it can be changed to DEBUG via the --log-level
	// command-line flag.
	LogOutput io.Writer
	// Optimization to avoid refreshing the OS package manager cache each time before
	// installing an OS package. Defaults to 1h.
	OsPkgCacheValidity time.Duration
	// The setup function, called before any command-line subcommand. No default.
	SetupFn func(prov *Provisioner) error
	// The configure function, called before the command-line configure subcommand.
	// No default.
	ConfigureFn ConfigureFn
}

// MainInt is a ready-made function for the main() of your installer.
//
// Usage:
//
//	func main() {
//		os.Exit(florist.MainInt(&florist.Options{
//			SetupFn:     setup,
//			ConfigureFn: configure,
//		}))
//	}
func MainInt(opts *Options) int {
	if err := MainErr(opts); err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

// MainErr is a ready-made function for the main() of your installer.
func MainErr(opts *Options) error {
	start := time.Now()

	if currentUser != nil {
		return fmt.Errorf("florist.Main: function already called")
	}

	if opts.LogOutput == nil {
		opts.LogOutput = os.Stdout
	}
	if opts.OsPkgCacheValidity == 0 {
		opts.OsPkgCacheValidity = time.Hour
	}
	if opts.SetupFn == nil {
		return fmt.Errorf("florist.Main: SetupFn is nil")
	}
	if opts.ConfigureFn == nil {
		return fmt.Errorf("florist.Main: ConfigureFn is nil")
	}

	var args cliArgs
	p := arg.MustParse(&args)
	if p.Subcommand() == nil {
		return fmt.Errorf("missing command")
	}

	if err := LowLevelInit(opts.LogOutput, args.LogLevel, opts.OsPkgCacheValidity); err != nil {
		return err
	}

	prov := newProvisioner()
	if err := opts.SetupFn(prov); err != nil {
		return fmt.Errorf("florist.Main: setup: %s", err)
	}

	log := Log()
	var err error
	switch {
	case args.List != nil:
		return args.List.Run(prov)
	case args.Install != nil:
		log.Info("starting", "command-line", os.Args)
		err = args.Install.Run(prov)
	case args.Configure != nil:
		log.Info("starting", "command-line", os.Args)
		err = args.Configure.Run(prov, opts.ConfigureFn)
	default:
		return fmt.Errorf("internal error: unwired command: %s", p.SubcommandNames())
	}

	elapsed := time.Since(start).Round(time.Millisecond)
	if err != nil {
		log.Error("exiting", "status", "failure", "error", err, "elapsed", elapsed)
		return err
	}
	log.Info("exiting", "status", "success", "elapsed", elapsed)
	return nil
}

type listArgs struct {
}

func (cmd *listArgs) Run(prov *Provisioner) error {
	for _, k := range prov.ordered {
		v := prov.flowers[k]
		if err := v.Init(); err != nil {
			return err
		}
		fmt.Printf("%s -- %s\n", v, v.Description())
		for _, fi := range v.Embedded() {
			fmt.Printf("  %s\n", fi)
		}
	}
	return nil
}

type installArgs struct{}

func (cmd *installArgs) Run(prov *Provisioner) error {
	log := Log()
	log.Info("installing", "flower-size", len(prov.flowers),
		"flowers", prov.ordered)

	for _, k := range prov.ordered {
		fl := prov.flowers[k]
		log.Info("installing", "flower", fl.String())
		if err := fl.Init(); err != nil {
			return fmt.Errorf("install: %s", err)
		}
		if err := fl.Install(); err != nil {
			return err
		}
	}

	return customizeMotd("installed", prov.root)
}

type configureArgs struct {
	Settings []string `help:"Settings and secrets file(s) (JSON)"`
}

func (cmd *configureArgs) Run(prov *Provisioner, configure ConfigureFn) error {
	log := Log()
	config, err := NewConfig(cmd.Settings)
	if err != nil {
		return fmt.Errorf("configure: %s", err)
	}

	if err := configure(prov, config); err != nil {
		return fmt.Errorf("configure: %s", err)
	}
	if err := config.Errors(); err != nil {
		return fmt.Errorf("configure: %s", err)
	}

	log.Info("configuring", "flowers-size", len(prov.flowers),
		"flowers", prov.ordered)

	for _, k := range prov.ordered {
		fl := prov.flowers[k]
		log.Info("configuring", "flower", fl.String())
		if err := fl.Init(); err != nil {
			return fmt.Errorf("configure: %s", err)
		}
		if err := fl.Configure(); err != nil {
			return fmt.Errorf("configure: %s", err)
		}
	}

	if err := customizeMotd("configured", prov.root); err != nil {
		return fmt.Errorf("configure: %s", err)
	}
	return nil
}

type Provisioner struct {
	flowers map[string]Flower
	ordered []string
	root    string
	errs    []string
}

func newProvisioner() *Provisioner {
	return &Provisioner{
		flowers: make(map[string]Flower),
	}
}

// FIXME what is this doing????
func (prov *Provisioner) UseWorkdir() {
	prov.root = WorkDir
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
	line := fmt.Sprintf("%s System %s by 🌼 florist 🌺\n", now, op)
	name := path.Join(rootDir, "/etc/motd")
	log.Debug("customize-motd", "target", name, "operation", op)

	if err := Mkdir(path.Dir(name), User().Username, 0755); err != nil {
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

// User returns the current user, as set by Init.
func User() *user.User {
	if currentUser == nil {
		panic("florist.User: must call florist.Main before")
	}
	return currentUser
}

func CacheValidity() time.Duration {
	if osPkgCacheValidity == 0 {
		panic("florist.CacheValidity: must call florist.Main before")
	}
	return osPkgCacheValidity
}

func Log() *slog.Logger {
	if floristLog == nil {
		panic("florist.Log: must call florist.Main before")
	}
	return floristLog
}

// LowLevelInit should be called only by low-level test code.
// Absolutely do not call in non-test code!
func LowLevelInit(logOutput io.Writer, logLevel string, cacheValidity time.Duration) error {
	var level slog.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		return fmt.Errorf("florist.Main: --log-level: %s", err)
	}

	prog := filepath.Base(os.Args[0])
	floristLog = slog.New(slog.NewTextHandler(logOutput,
		&slog.HandlerOptions{Level: level})).With("prog", prog, "lib", "florist")

	// FIXME should this go below???
	var err error
	currentUser, err = user.Current()
	if err != nil {
		return fmt.Errorf("florist.Main: %s", err)
	}

	osPkgCacheValidity = cacheValidity

	return nil
}
