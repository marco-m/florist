package apt_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/marco-m/florist/pkg/apt"
	"github.com/marco-m/florist/pkg/florist"
)

/*
TODO Optimization
If Florist is doing an apt.AddRepo, it should Update only ONCE in lifetime.
I can begin by recording if apt.AddRepo has been called after an apt.update and
print a warning by saying that the flowers should be re-ordered to first call
apt.addRepo and then apt.Install !

*/

func TestAptInstallAndRemove(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	if err := apt.Install("ripgrep"); err != nil {
		t.Errorf("apt.Install: %s", err)
	}

	if err := apt.Remove("ripgrep"); err != nil {
		t.Errorf("apt.Remove: %s", err)
	}
}

func TestAddRepo(t *testing.T) {
	florist.SkipIfNotDisposableHost(t)

	if err := apt.AddRepo(
		"docker",
		"https://download.docker.com/linux/debian/gpg",
		"1500c1f56fa9e26b9b8f42452a553675796ade0807cdce11975eb98170b3a570",
		"https://download.docker.com/linux/debian",
	); err != nil {
		t.Errorf("apt.AddRepo: %s", err)
	}

	repoFile := "/etc/apt/sources.list.d/docker.list"
	buf, err := os.ReadFile(repoFile)
	if err != nil {
		t.Fatal("reading generated file:", err)
	}
	have := string(bytes.TrimSpace(buf))
	// Full string varies per OS version. For example:
	// deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian noble stable
	// deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian bookworm stable
	want := "deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian"
	if !strings.Contains(have, want) {
		t.Errorf("\ngenerated file\nhave: %s\ndoes not contain: %s", have, want)
	}
}
