// Package taskfile installs the Task utility
package taskfile

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/marco-m/florist/pkg/florist"
)

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	fsys    fs.FS
	Version string
	Hash    string
	log     hclog.Logger
}

func (fl Flower) String() string {
	return "taskfile"
}

func (fl Flower) Description() string {
	return "install the Task (simpler make alternative) utility"
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

func (fl *Flower) Install(files fs.FS, finder florist.Finder) error {
	fl.log.Info("begin")
	defer fl.log.Info("end")

	taskexe := "/usr/local/bin/task"
	if foundTaskVersion :=
		maybeTaskfileVersion(fl.log, taskexe); foundTaskVersion == fl.Version {
		fl.log.Debug("Task already installed with matching version", "version",
			fl.Version)
		return nil
	}

	fl.log.Info("Download task package", "version", fl.Version)
	// https://github.com/go-task/task/releases/download/v3.9.0/task_linux_amd64.tar.gz
	url := fmt.Sprintf("https://github.com/go-task/task/releases/download/v%s/task_linux_amd64.tar.gz", fl.Version)
	client := &http.Client{Timeout: 30 * time.Second}
	tgzPath, err := florist.NetFetch(client, url, florist.SHA256, fl.Hash, florist.WorkDir)
	if err != nil {
		return fmt.Errorf("%s: %s", fl.log.Name(), err)
	}

	fl.log.Debug("extracting Task")
	if err := os.RemoveAll(path.Join(florist.WorkDir, "task")); err != nil {
		return fmt.Errorf("%s: %s", fl.log.Name(), err)
	}
	dst := path.Join(florist.WorkDir, "task")
	if err := os.Mkdir(dst, 0755); err != nil {
		return fmt.Errorf("%s: %s", fl.log.Name(), err)
	}
	cmd := exec.Command("tar", "xzf", tgzPath)
	cmd.Dir = dst
	if err := florist.CmdRun(fl.log, cmd); err != nil {
		return fmt.Errorf("%s: %s", fl.log.Name(), err)
	}

	fl.log.Debug("Moving Task into place")
	if err := os.Rename(path.Join(florist.WorkDir, "task", "task"), taskexe); err != nil {
		return fmt.Errorf("%s: %s", fl.log.Name(), err)
	}
	fl.log.Info("Installed Task", "path", taskexe)

	return nil
}

func (fl *Flower) Configure(files fs.FS, finder florist.Finder) error {
	return nil
}

// maybeGoVersion returns the version such as "1.17.2" if found, or the empty
// string if not found.
func maybeTaskfileVersion(log hclog.Logger, taskexe string) string {
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
