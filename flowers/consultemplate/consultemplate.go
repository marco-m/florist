package consultemplate

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/creasty/defaults"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

//go:embed embedded
var embedded embed.FS

const (
	HomeDir      = "/opt/consul-template"
	ConfigDir    = HomeDir + "/config"
	TemplatesDir = HomeDir + "/templates"
	BinDir       = "/usr/local/bin"

	// relative to the Go embed FS
	ConfigDirSrc    = "embedded/config"
	TemplatesDirSrc = "embedded/templates"
	UnitFile        = "embedded/consul-template.service"
)

const Name = "consul-template"

var _ florist.Flower = (*Flower)(nil)

type Flower struct {
	Inst
	Conf
}

type Inst struct {
	Version            string
	Hash               string
	ConfigurationFiles []string
	TemplateFiles      []string
	Fsys               fs.FS
}

type Conf struct{}

func (fl *Flower) String() string {
	return Name
}

func (fl *Flower) Description() string {
	return "install " + Name
}

func (fl *Flower) Embedded() []string {
	return nil
}

func (fl *Flower) Init() error {
	if fl.Fsys == nil {
		fl.Fsys = embedded
	}
	if err := defaults.Set(fl); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}
	if fl.Version == "" {
		return fmt.Errorf("%s: %s", Name, "Version cannot be empty")
	}
	if fl.Hash == "" {
		return fmt.Errorf("%s: %s", Name, "Hash cannot be empty")
	}
	return nil
}

func (fl *Flower) Install() error {
	log := slog.With("flower", Name+".install")

	log.Info("Add system user 'consul-template'")
	if err := florist.UserAdd("consul-template", florist.UserAddOpts{
		System:  true,
		HomeDir: HomeDir,
	}); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	log.Info("Create consul-template configuration dir", "dir", ConfigDir)
	if err := os.MkdirAll(ConfigDir, 0o755); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	log.Info("Create consul-template templates dir", "dir", TemplatesDir)
	if err := os.MkdirAll(TemplatesDir, 0o755); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	// FIXME SECURITY TODO
	//
	// For the time being we run consul-template as root :-(
	//
	// Talking with MA, the proper way to use consul-template is to use it in
	// supervisor mode for each service that needs it (so for each service, the systemd
	// unit file instead of starting the service, starts a dedicated consul-template),
	// so that we can avoid having it running as root!

	if err := installExe(log, fl.Version, fl.Hash, "root"); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	for _, cfg := range fl.ConfigurationFiles {
		src := path.Join(ConfigDirSrc, cfg)
		dst := path.Join(ConfigDir, cfg)
		log.Info("Install consul-template configuration file", "dst", dst)
		if err := florist.CopyFileFs(fl.Fsys, src, dst, 0o640, "root"); err != nil {
			return fmt.Errorf("%s: %s", Name, err)
		}
	}

	for _, tmpl := range fl.TemplateFiles {
		src := path.Join(TemplatesDirSrc, tmpl)
		dst := path.Join(TemplatesDir, tmpl)
		log.Info("Install consul-template template file", "dst", dst)
		if err := florist.CopyFileFs(fl.Fsys, src, dst, 0o640, "root"); err != nil {
			return fmt.Errorf("%s: %s", Name, err)
		}
	}

	dst := path.Join("/etc/systemd/system/", filepath.Base(UnitFile))
	log.Info("Install consul-template systemd unit file", "dst", dst)
	if err := florist.CopyFileFs(fl.Fsys, UnitFile, dst, 0o644, "root"); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	log.Info("Enable service to start at boot", "unit", filepath.Base(UnitFile))
	if err := systemd.Enable(filepath.Base(UnitFile)); err != nil {
		return fmt.Errorf("%s: %s", Name, err)
	}

	// We do not start the service at Packer time because it is not needed.

	return nil
}

func (fl *Flower) Configure() error {
	return nil
}

func installExe(
	log *slog.Logger,
	version string,
	hash string,
	owner string,
) error {
	log.Info("Download consul-template package")
	url := fmt.Sprintf("https://releases.hashicorp.com/consul-template/%s/consul-template_%s_linux_amd64.zip", version, version)
	client := &http.Client{Timeout: 30 * time.Second}
	zipPath, err := florist.NetFetch(client, url, florist.SHA256, hash,
		florist.WorkDir)
	if err != nil {
		return err
	}

	extracted := path.Join(florist.WorkDir, "consul-template")
	log.Info("Unzipping consul-template package", "dst", extracted)
	if err := florist.UnzipOne(zipPath, "consul-template",
		extracted); err != nil {
		return err
	}

	exe := path.Join(BinDir, "consul-template")
	log.Info("Install consul-template", "dst", exe)
	if err := florist.CopyFile(extracted, exe, 0o755, owner); err != nil {
		return err
	}

	return nil
}
