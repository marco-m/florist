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
)

const (
	GoHome = "/usr/local/go"
	GoBin  = GoHome + "/bin"
)

type Options struct {
	Version string
	Hash    string
}

type Flower struct {
	Options
	log hclog.Logger
}

func New(opts Options) (*Flower, error) {
	fl := Flower{Options: opts}
	name := fmt.Sprintf("florist.flower.%s", fl)
	fl.log = florist.Log.ResetNamed(name)

	if fl.Version == "" {
		return nil, fmt.Errorf("%s.new: missing version", name)
	}
	if fl.Hash == "" {
		return nil, fmt.Errorf("%s.new: missing hash", name)
	}

	return &fl, nil
}

func (fl Flower) String() string {
	return "golang"
}

func (fl Flower) Description() string {
	return "install the Go programming language"
}

func (fl *Flower) Install() error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	goexe := path.Join(GoBin, "go")
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
		return fmt.Errorf("goRun: %s", err)
	}
	cmd := exec.Command("tar", "xzf", tgzPath)
	cmd.Dir = florist.WorkDir
	if err := florist.LogRun(fl.log, cmd); err != nil {
		return fmt.Errorf("goRun: %s", err)
	}

	fl.log.Debug("removing old Go (if any)")
	if err := os.RemoveAll(GoHome); err != nil {
		return fmt.Errorf("goRun: %s", err)
	}
	fl.log.Debug("Moving Go into place")
	if err := os.Rename(path.Join(florist.WorkDir, "go"), GoHome); err != nil {
		return fmt.Errorf("goRun: %s", err)
	}
	fl.log.Info("Installed Go", "path", GoHome)

	//
	// The never-ending pain of shells and something apparently as simple as
	// adding to the PATH environment variable. Sigh.
	//

	// Works for POSIX shells but not for fish
	dst := "/etc/profile.d/go.sh"
	fl.log.Debug("Adding Go to PATH (POSIX shells)", "dst", dst)
	fi, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("goRun: %s", err)
	}
	line := fmt.Sprintf("export PATH=$PATH:%s\n", GoBin)
	if _, err := fi.WriteString(line); err != nil {
		return fmt.Errorf("goRun: write to %s: %s", dst, err)
	}

	// Works for fish
	if _, err := os.Stat("/etc/fish/conf.d"); err == nil {
		dst := "/etc/fish/conf.d/go.fish"
		fl.log.Debug("Adding Go to PATH (Fish shell)", "dst", dst)
		fi, err := os.Create(dst)
		if err != nil {
			return fmt.Errorf("goRun: %s", err)
		}
		line := fmt.Sprintf("set -x PATH $PATH %s\n", GoBin)
		if _, err := fi.WriteString(line); err != nil {
			return fmt.Errorf("goRun: write to %s: %s", dst, err)
		}
	}
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
