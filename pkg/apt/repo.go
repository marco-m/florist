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

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/platform"
)

// AddRepo securely adds an APT repo with corresponding PGP key.
// NOTE After having called AddRepo, you MUST call [Refresh] for the new repo
// to be seen by a subsequent [Install].
//
// See: https://wiki.debian.org/DebianRepository/UseThirdParty
//
// Example:
//
//	if err := apt.AddRepo(
//		"docker",
//		"https://download.docker.com/linux/debian",
//		"https://download.docker.com/linux/debian/gpg",
//		"1500c1f56fa9e26b9b8f42452a553675796ade0807cdce11975eb98170b3a570",
//	); err != nil {
//		return err
//	}
//	if err := apt.Update(0 * time.Second); err != nil {
//		return err
//	}
func AddRepo(
	name string,
	keyURL string,
	keyHash string,
	repoURL string,
) error {
	const fn = "apt.AddRepo"
	log := slog.With("fn", fn)

	log.Info("Download PGP key", "url", keyURL)
	client := &http.Client{Timeout: 15 * time.Second}
	keyPath, err := florist.NetFetch(client, keyURL, florist.SHA256, keyHash, florist.WorkDir)
	if err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}

	const keyringsDir = "/etc/apt/keyrings"
	keyDst := filepath.Join(keyringsDir, "docker.asc")
	if err := florist.Mkdir(keyringsDir, 0o755, "root", "root"); err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}

	log.Info("Install PGP key", "dst", keyDst)
	if err := florist.CopyFile(keyPath, keyDst, 0o644, "root"); err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}

	arch := runtime.GOARCH
	osInfo, err := platform.CollectInfo()
	if err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}

	repoListDir := "/etc/apt/sources.list.d/"
	repoListPath := path.Join(repoListDir, name+".list")
	if err := os.MkdirAll(repoListDir, 0o755); err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}

	repoLine := fmt.Sprintf("deb [arch=%s signed-by=%s] %s %s stable\n",
		arch, keyDst, repoURL, osInfo.Codename)
	log.Debug("write file", "repoline", repoLine, "repoListPath", repoListPath)
	if err := florist.WriteFile(repoListPath, repoLine, 0o644, "root", "root"); err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}

	return nil
}
