// Package golang contains a flower to install the Go programming language
package golang

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/pkg/envvar"
	"github.com/marco-m/florist/pkg/florist"
)

const (
	GOROOT = "/usr/local/go"
)

const Name = "golang"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	Version string
	Hash    string
}

type Conf struct{}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "install the Go programming language"
}

func (fl *Flower) Embedded() []string {
	return nil
}

func (fl *Flower) Init() error {
	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	if fl.Version == "" {
		return fmt.Errorf("%s.new: missing version", Name)
	}
	if fl.Hash == "" {
		return fmt.Errorf("%s.new: missing hash", Name)
	}
	return nil
}

func (fl *Flower) Install() error {
	log := florist.Log().With("flower", Name+".install")

	goexe := path.Join(GOROOT, "bin/go")
	if installedGoVersion(log, goexe) == fl.Version {
		log.Debug("Go already installed with matching version", "version", fl.Version)
		return nil
	}

	log.Info("Download Go package", "version", fl.Version)
	uri, err := url.JoinPath("https://golang.org/dl",
		"go"+fl.Version+".linux-amd64.tar.gz")
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	tgzPath, err := florist.NetFetch(client, uri, florist.SHA256, fl.Hash,
		florist.WorkDir)
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	log.Debug("extracting Go")
	if err := os.RemoveAll(path.Join(florist.WorkDir, "go")); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	cmd := exec.Command("tar", "xzf", tgzPath)
	cmd.Dir = florist.WorkDir
	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	log.Debug("removing old Go (if any)")
	if err := os.RemoveAll(GOROOT); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	log.Debug("Moving Go into place")
	if err := os.Rename(path.Join(florist.WorkDir, "go"), GOROOT); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	log.Debug("Creating symbolic links")
	goBinDir := path.Join(GOROOT, "bin")
	goBinaries, err := os.ReadDir(goBinDir)
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	for _, de := range goBinaries {
		oldname := path.Join(goBinDir, de.Name())
		newname := path.Join("/usr/local/bin", de.Name())
		// Remove old Go symlink (if any). If no previous Go installation, the symlink
		// will not exist and os.Remove will return an error, so we don't check for it.
		os.Remove(newname)
		if err := os.Symlink(oldname, newname); err != nil {
			return fmt.Errorf("%s: %s", Name, err)
		}
	}

	// Smoke test for the installation.
	goExe := filepath.Join(goBinDir, "go")
	output, err := exec.Command(goExe, "version").Output()
	if err != nil {
		return fmt.Errorf("%s: doing Go install smoke test: %s", Name, err)
	}
	log.Info("Installed Go", "path", GOROOT, "smoke", string(bytes.TrimSpace(output)))

	return envvar.AddPaths(log, "go", "$HOME/go/bin")
}

func (fl *Flower) Configure() error {
	log := florist.Log().With("flower", Name+".configure")
	log.Debug("nothing to do")
	return nil
}

// installedGoVersion returns the version such as "1.17.2" if found, or the empty
// string if not found.
func installedGoVersion(log *slog.Logger, goexe string) string {
	log = log.With("path", goexe)
	goexe, err := exec.LookPath(goexe)
	if err != nil {
		log.Debug("Go executable not found")
		return ""
	}
	cmd := exec.Command(goexe, "version")
	out, err := cmd.Output()
	if err != nil {
		log.Debug("unexpected", "err", err)
		return ""
	}
	// $ go version
	// go version go1.17.2 darwin/amd64
	// 0  1       2        3
	tokens := strings.Fields(string(out))
	if len(tokens) != 4 {
		log.Debug("unexpected", "go version", tokens)
		return ""
	}
	version := strings.TrimPrefix(tokens[2], "go")
	log.Debug("Found Go", "version", version)
	return version
}
