package florist

import (
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/marco-m/florist/internal"
)

// Garden is passed to each invocation of [Flower.Init].
type Garden struct {
	logLevel slog.Level // settable from the CLI --log-level
	//
	user    *user.User
	group   *user.Group
	workDir string
	//
	start time.Time
	log   *slog.Logger
	prov  *Provisioner
	opts  *Options
}

// NewGarden should be called only by low-level test code.
// Absolutely do not call in non-test code! Call florist.MainInt instead!
func NewGarden(opts *Options) (*Garden, error) {
	const fn = "florist.NewGarden"
	errorf := internal.MakeErrorf(fn)

	if opts.LogOutput == nil {
		opts.LogOutput = os.Stdout
	}
	if opts.TempDir == "" {
		opts.TempDir = os.TempDir()
	}
	if opts.RootDir == "" {
		// TODO This should have a sane default per OS.
		opts.RootDir = "/"
	}

	gdn := &Garden{
		start:   time.Now(),
		prov:    newProvisioner(),
		workDir: filepath.Join(opts.TempDir, "florist"),
		opts:    opts,
	}

	prog := filepath.Base(os.Args[0])
	slog.SetDefault(slog.New(slog.NewTextHandler(opts.LogOutput,
		&slog.HandlerOptions{Level: opts.LogLevel})).With("florist", prog))
	gdn.log = slog.Default()

	var err error
	gdn.user, err = user.Current()
	if err != nil {
		return nil, errorf("%s", err)
	}
	gdn.group, err = user.LookupGroupId(gdn.user.Gid)
	if err != nil {
		return nil, errorf("%s", err)
	}

	if err := os.MkdirAll(opts.TempDir, 0o755); err != nil {
		return nil, errorf("os.MkdirAll: %s", err)
	}
	gdn.log.Debug(fn, "op", "MkdirAll-TempDir", "path", opts.TempDir)

	if err := Mkdir(gdn.workDir, 0o755, gdn.User(), gdn.Group()); err != nil {
		return nil, errorf("%s", err)
	}
	gdn.log.Debug(fn, "op", "Mkdir-WorkDir", "path", gdn.workDir)

	return gdn, nil
}

// User returns the current user name.
func (gdn Garden) User() string {
	return gdn.user.Username
}

// Group returns the primary group name of the current user.
func (gdn Garden) Group() string {
	return gdn.group.Name
}

func (gdn Garden) WorkDir() string {
	return gdn.workDir
}
