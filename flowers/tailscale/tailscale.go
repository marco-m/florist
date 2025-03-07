// Package tailscale contains a flower to install and configure Tailscale.
package tailscale

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/creasty/defaults"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

// From https://pkgs.tailscale.com/stable/#static, we get for example
// https://pkgs.tailscale.com/stable/tailscale_1.80.0_amd64.tgz
// It contains 2 executables and some configuration files.
// - tailscale
// - tailscaled

//go:embed embedded
var embedded embed.FS

const (
	BinDir      = "/usr/local/bin"
	SbinDir     = "/usr/sbin"
	DefaultsDir = "/etc/default"

	// relative to the Go embed FS
	UnitFile     = "embedded/tailscaled.service"
	DefaultsFile = "embedded/tailscaled.defaults"
	//
	DefaultsFileDst = "/etc/default/tailscaled"
)

const Name = "tailscale"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	Version string
	Hash    string
	Fsys    fs.FS
}

type Conf struct {
	// https://tailscale.com/kb/1085/auth-keys
	AuthKey string
	// https://tailscale.com/kb/1193/tailscale-ssh
	Ssh bool
}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "install " + Name
}

func (fl *Flower) Embedded() []string {
	return nil
}

func (fl *Flower) Init() error {
	errorf := makeErrorf(Name + ".init")
	if fl.Fsys == nil {
		fl.Fsys = embedded
	}
	if err := defaults.Set(fl); err != nil {
		return errorf("%s", err)
	}
	if fl.Version == "" {
		return errorf("version cannot be empty")
	}
	if fl.Hash == "" {
		return errorf("hash cannot be empty")
	}
	return nil
}

func (fl *Flower) Install() error {
	errorf := makeErrorf(Name + ".install")
	log := florist.Log().With("flower", Name+".install")

	if err := installExes(log, fl.Version, fl.Hash, "root"); err != nil {
		return errorf("%s", err)
	}

	dst := path.Join("/etc/systemd/system/", filepath.Base(UnitFile))
	log.Info("Install tailscale systemd unit file", "dst", dst)
	if err := florist.CopyFileFs(fl.Fsys, UnitFile, dst, 0o644, "root"); err != nil {
		return errorf("%s", err)
	}

	log.Info("install-environment-file", "dst", dst)
	if err := florist.CopyFileFs(fl.Fsys, DefaultsFile, DefaultsFileDst, 0o644, "root"); err != nil {
		return errorf("%s", err)
	}

	log.Info("Enable service to start at boot", "unit", filepath.Base(UnitFile))
	if err := systemd.Enable(filepath.Base(UnitFile)); err != nil {
		return errorf("%s", err)
	}

	return nil
}

func (fl *Flower) Configure() error {
	errorf := makeErrorf(Name + ".configure")
	log := florist.Log().With("flower", Name+".configure")

	keyFile := filepath.Join(florist.WorkDir, Name, "authkey")
	defer func() {
		if err := os.Remove(keyFile); err != nil {
			log.Error("remove-authkey-file", "path", keyFile, "err", err)
		} else {
			log.Debug("remove-authkey-file", "path", keyFile)
		}
	}()
	log.Debug("write-authkey-file", "path", keyFile)
	if err := florist.WriteFile(keyFile, fl.AuthKey, 0o600, "root"); err != nil {
		return errorf("writing the authkey file: %s", err)
	}

	if err := systemd.Start(filepath.Base(UnitFile)); err != nil {
		return errorf("%s", err)
	}

	if err := florist.CmdRun(log, exec.Command("tailscale", "version")); err != nil {
		return errorf("printing tailscale version: %s", err)
	}

	//
	// Before doing any customization with "tailscale set", we must first call
	// "tailscale up"
	//

	log.Info("tailscale-up")
	// --auth-key string
	//     Node authorization key; if it begins with "file:", then it's a path
	//     to a file containing the authkey.
	upArgs := []string{"up", "--auth-key=file:" + keyFile}
	if fl.Ssh {
		upArgs = append(upArgs, "--ssh")
	}
	if err := florist.CmdRun(log, exec.Command("tailscale", upArgs...)); err != nil {
		errorf("tailscale up: %s", err)
	}

	if err := florist.CmdRun(log, exec.Command("tailscale", "set", "--auto-update")); err != nil {
		return errorf("setting tailscale auto-update: %s", err)
	}

	return nil
}

func installExes(
	log *slog.Logger,
	version string,
	hash string,
	owner string,
) error {
	log.Info("Download tailscale package")
	url := fmt.Sprintf("https://pkgs.tailscale.com/stable/tailscale_%s_amd64.tgz", version)
	client := &http.Client{Timeout: 30 * time.Second}
	tarPath, err := florist.NetFetch(client, url, florist.SHA256, hash,
		florist.WorkDir)
	if err != nil {
		return err
	}

	workdir := path.Join(florist.WorkDir, Name)
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		return err
	}

	//
	// tailscaled
	//

	name := "tailscaled"
	extracted := path.Join(workdir, name)
	log.Info("unarchive", "name", name, "dst", extracted)
	if err := florist.UntarOne(tarPath, name, extracted); err != nil {
		return err
	}
	dst := path.Join(SbinDir, name)
	log.Info("install", "name", extracted, "dst", dst)
	if err := florist.CopyFile(extracted, dst, 0o755, owner); err != nil {
		return err
	}

	//
	// tailscale
	//

	name = "tailscale"
	extracted = path.Join(workdir, name)
	log.Info("unarchive", "name", name, "dst", extracted)
	if err := florist.UntarOne(tarPath, name, extracted); err != nil {
		return err
	}
	dst = path.Join(BinDir, name)
	log.Info("install", "name", extracted, "dst", dst)
	if err := florist.CopyFile(extracted, dst, 0o755, owner); err != nil {
		return err
	}

	return nil
}

func makeErrorf(prefix string) func(format string, a ...any) error {
	return func(format string, a ...any) error {
		return fmt.Errorf(prefix+": "+format, a...)
	}
}
