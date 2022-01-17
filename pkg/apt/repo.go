package apt

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/marco-m/florist"
)

// Securely add a APT repo with corresponding PGP key.
// https://wiki.debian.org/DebianRepository/UseThirdParty
func AddRepo(
	name string,
	repoURL string,
	keyURL string,
	keyHash string,
) error {
	log := florist.Log.Named("apt.AddRepo")

	log.Info("Install packages needed to add a repo")
	if err := Install(
		"gpg",
	); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("Download PGP key", "url", keyURL)
	client := &http.Client{Timeout: 15 * time.Second}
	keyPath, err := florist.NetFetch(client, keyURL, florist.SHA256, keyHash, florist.WorkDir)
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	keyDst := fmt.Sprintf("/usr/share/keyrings/%s-archive-keyring.gpg", name)
	log.Info("Install PGP key", "dst", keyDst)
	// delete key file if present, to avoid gpg going into interactive mode
	os.Remove(keyDst)
	cmd := exec.Command("gpg", "--dearmor", "-o", keyDst, keyPath)
	if err := florist.LogRun(log, cmd); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	arch, err := exec.Command("dpkg", "--print-architecture").Output()
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}
	codename, err := exec.Command("lsb_release", "-cs").Output()
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	repoLine := fmt.Sprintf("deb [arch=%s signed-by=%s] %s %s stable\n",
		string(bytes.TrimSpace(arch)), keyDst, repoURL, string(bytes.TrimSpace(codename)))
	repoListPath := fmt.Sprintf("/etc/apt/sources.list.d/%s.list", name)

	// write repoline to repolistpath
	log.Debug("write file", "repoline", repoLine, "repoListPath", repoListPath)
	fi, err := os.Create(repoListPath)
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}
	defer fi.Close()
	if _, err := fi.WriteString(repoLine); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	return nil
}
