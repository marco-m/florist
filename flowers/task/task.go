// Package task installs the Task utility
package task

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/creasty/defaults"
	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/pkg/florist"
)

const Name = "task"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	Version string
	Hash    string
}

type Conf struct {
}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "install the Task (simpler make alternative) utility"
}

func (fl *Flower) Embedded() []string {
	return nil
}

func (fl *Flower) Init() error {
	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	if fl.Version == "" {
		return fmt.Errorf("%s: missing Version", Name)
	}
	if fl.Hash == "" {
		return fmt.Errorf("%s: missing Hash", Name)
	}
	return nil
}

func (fl *Flower) Install() error {
	log := florist.Log.ResetNamed(Name + ".install")

	taskDst := "/usr/local/bin/task"
	if installedTaskVersion(log, taskDst) == fl.Version {
		log.Debug("Task already installed with matching version", "version",
			fl.Version)
		return nil
	}

	log.Info("downloading Task", "version", fl.Version)
	// https://github.com/go-task/task/releases/download/v3.9.0/task_linux_amd64.tar.gz
	uri, err := url.JoinPath("https://github.com/go-task/task/releases/download/",
		"v"+fl.Version, "task_linux_amd64.tar.gz")
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	tgzPath, err := florist.NetFetch(client, uri, florist.SHA256, fl.Hash,
		florist.WorkDir)
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	dstDir, err := os.MkdirTemp(florist.WorkDir, Name)
	if err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	log.Debug("extracting Task", "dir", dstDir)
	cmd := exec.Command("tar", "xzf", tgzPath)
	cmd.Dir = dstDir
	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	log.Debug("Moving Task into place")
	taskSrc := path.Join(dstDir, "task")
	if err := os.Rename(taskSrc, taskDst); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	log.Info("Installed Task", "path", taskDst)

	// We remove the tmpdir only on success; in case of failure we leave
	// it around, it can be useful to troubleshoot.
	if err := os.RemoveAll(dstDir); err != nil {
		log.Info("removing tmpdir", "dir", dstDir, "err", err)
	}

	return nil
}

func (fl *Flower) Configure() error {
	log := florist.Log.ResetNamed(Name + ".configure")
	log.Debug("nothing to do")
	return nil
}

// installedTaskVersion returns the version such as "1.17.2" if found, or the empty
// string if not found.
func installedTaskVersion(log hclog.Logger, taskexe string) string {
	log = log.With("path", taskexe)
	taskexe, err := exec.LookPath(taskexe)
	if err != nil {
		log.Debug("Task executable not found")
		return ""
	}
	cmd := exec.Command(taskexe, "--version")
	out, err := cmd.Output()
	if err != nil {
		log.Debug("unexpected", "err", err)
		return ""
	}
	// $ task --version
	// Task version: v3.9.0 (h1:k6ZHDqKU+NCJxN2PhHRB4gF0NSZVlRdRSbtP89pDKDA=)
	// 0    1        2      3
	tokens := strings.Fields(string(out))
	if len(tokens) < 3 {
		log.Debug("unexpected", "task version", tokens)
		return ""
	}
	version := strings.TrimPrefix(tokens[2], "v")
	log.Debug("Found Task", "version", version)
	return version
}
