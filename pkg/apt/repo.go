package apt

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/marco-m/florist/pkg/florist"
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
	log := florist.Log().With("fn", fn)

	log.Info("Install packages needed to add a repo")
	if err := Install(
		"gpg",
	); err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}

	log.Info("Download PGP key", "url", keyURL)
	client := &http.Client{Timeout: 15 * time.Second}
	keyPath, err := florist.NetFetch(client, keyURL, florist.SHA256, keyHash, florist.WorkDir)
	if err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}

	keyDst := fmt.Sprintf("/usr/share/keyrings/%s-archive-keyring.gpg", name)
	log.Info("Install PGP key", "dst", keyDst)
	// delete key file if present, to avoid gpg going into interactive mode
	os.Remove(keyDst)
	cmd := exec.Command("gpg", "--dearmor", "-o", keyDst, keyPath)
	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}

	arch, err := exec.Command("dpkg", "--print-architecture").Output()
	if err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}
	codename, err := exec.Command("lsb_release", "-cs").Output()
	if err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}

	repoListDir := "/etc/apt/sources.list.d/"
	if err := os.MkdirAll(repoListDir, 0o755); err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}
	repoLine := fmt.Sprintf("deb [arch=%s signed-by=%s] %s %s stable\n",
		string(bytes.TrimSpace(arch)), keyDst, repoURL, string(bytes.TrimSpace(codename)))
	repoListPath := path.Join(repoListDir, name+".list")
	// write repoline to repolistpath
	log.Debug("write file", "repoline", repoLine, "repoListPath", repoListPath)
	fi, err := os.Create(repoListPath)
	if err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}
	defer fi.Close()
	if _, err := fi.WriteString(repoLine); err != nil {
		return fmt.Errorf("%s: %s", fn, err)
	}

	return nil
}
