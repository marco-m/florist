package apt

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/marco-m/florist/internal"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/platform"
)

// AddRepo securely adds an APT repo with corresponding PGP key. NOTE After having called
// AddRepo, there is no need to refresh the APT cache: a subsequent [Install] will detect
// that a new repo has been added and will call "apt update" behind the scenes.
//
// See: https://wiki.debian.org/DebianRepository/UseThirdParty
//
// Example:
//
//	if err := apt.AddRepo(
//		"docker",
//		"https://download.docker.com/linux/debian/gpg",
//		"1500c1f56fa9e26b9b8f42452a553675796ade0807cdce11975eb98170b3a570",
//		"https://download.docker.com/linux/debian",
//	); err != nil {
//		return err
//	}
func AddRepo(name string, keyURL string, keyHash string, repoURL string) error {
	errorf, log := internal.MakeErrorfAndLog("apt.AddRepo", slog.Default())

	log.Info("Download PGP key", "url", keyURL)
	client := &http.Client{Timeout: 15 * time.Second}
	keyPath, err := florist.NetFetch(client, keyURL, florist.SHA256, keyHash, florist.WorkDir)
	if err != nil {
		return errorf("%s", err)
	}

	const keyringsDir = "/etc/apt/keyrings"
	keyDst := filepath.Join(keyringsDir, "docker.asc")
	if err := florist.Mkdir(keyringsDir, 0o755, "root", "root"); err != nil {
		return errorf("%s", err)
	}

	log.Info("Install PGP key", "dst", keyDst)
	if err := florist.CopyFile(keyPath, keyDst, 0o644, "root"); err != nil {
		return errorf("%s", err)
	}

	arch := runtime.GOARCH
	osInfo, err := platform.CollectInfo()
	if err != nil {
		return errorf("%s", err)
	}

	repoListDir := "/etc/apt/sources.list.d/"
	repoListPath := path.Join(repoListDir, name+".list")
	if err := os.MkdirAll(repoListDir, 0o755); err != nil {
		return errorf("%s", err)
	}

	repoLine := fmt.Sprintf("deb [arch=%s signed-by=%s] %s %s stable\n",
		arch, keyDst, repoURL, osInfo.Codename)
	log.Debug("write file", "repoline", repoLine, "repoListPath", repoListPath)
	if err := florist.WriteFile(repoListPath, repoLine, 0o644, "root", "root"); err != nil {
		return errorf("%s", err)
	}

	if err := cacheState.Invalidate(); err != nil {
		return errorf("%s", err)
	}

	return nil
}
