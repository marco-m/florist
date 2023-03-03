// Package golang contains a flower to install the Go programming language
package golang

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/envpath"
)

const (
	Goroot = "/usr/local/go"
)

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Version string
	Hash    string
	log     hclog.Logger
}

func (fl *Flower) String() string {
	return "golang"
}

func (fl *Flower) Description() string {
	return "install the Go programming language"
}

func (fl *Flower) Init() error {
	name := fmt.Sprintf("florist.flower.%s", fl)
	fl.log = florist.Log.ResetNamed(name)

	if fl.Version == "" {
		return fmt.Errorf("%s.new: missing version", name)
	}
	if fl.Hash == "" {
		return fmt.Errorf("%s.new: missing hash", name)
	}

	return nil
}

func (fl *Flower) Install() error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	goexe := path.Join(Goroot, "bin/go")
	if foundGoVersion := maybeGoVersion(fl.log, goexe); foundGoVersion == fl.Version {
		fl.log.Debug("Go already installed with matching version", "version", fl.Version)
		return nil
	}

	fl.log.Info("Download Go package", "version", fl.Version)
	url := "https://" + path.Join("golang.org/dl",
		fmt.Sprintf("go%s.linux-amd64.tar.gz", fl.Version))
	client := &http.Client{Timeout: 30 * time.Second}
	tgzPath, err := florist.NetFetch(client, url, florist.SHA256, fl.Hash, florist.WorkDir)
	if err != nil {
		return err
	}

	fl.log.Debug("extracting Go")
	if err := os.RemoveAll(path.Join(florist.WorkDir, "go")); err != nil {
		return fmt.Errorf("go.Install: %s", err)
	}
	cmd := exec.Command("tar", "xzf", tgzPath)
	cmd.Dir = florist.WorkDir
	if err := florist.CmdRun(fl.log, cmd); err != nil {
		return fmt.Errorf("go.Install: %s", err)
	}

	fl.log.Debug("removing old Go (if any)")
	if err := os.RemoveAll(Goroot); err != nil {
		return fmt.Errorf("go.Install: %s", err)
	}
	// TODO iterate in go/bin/ and remove symlinks
	fl.log.Debug("Moving Go into place")
	if err := os.Rename(path.Join(florist.WorkDir, "go"), Goroot); err != nil {
		return fmt.Errorf("go.Install: %s", err)
	}
	fl.log.Debug("Creating symbolic links")
	binDir := path.Join(Goroot, "bin")
	dirEntries, err := os.ReadDir(binDir)
	if err != nil {
		return fmt.Errorf("go.Install: %s", err)
	}
	for _, de := range dirEntries {
		if err := os.Symlink(path.Join(binDir, de.Name()),
			path.Join("/usr/local/bin", de.Name())); err != nil {
			return fmt.Errorf("go.Install: %s", err)
		}
	}
	fl.log.Info("Installed Go", "path", Goroot)

	return envpath.Add(fl.log, "go", "$HOME/go/bin")
}

func (fl *Flower) Configure() error {
	return nil
}

// maybeGoVersion returns the version such as "1.17.2" if found, or the empty
// string if not found.
func maybeGoVersion(log hclog.Logger, goexe string) string {
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
