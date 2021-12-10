package consultemplate

import (
	"fmt"
	"io/fs"
	"net/http"
	"os/user"
	"path"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/marco-m/florist"
	"github.com/marco-m/florist/pkg/systemd"
)

const (
	HomeDir      = "/opt/consul-template"
	ConfigDir    = HomeDir + "/config"
	TemplatesDir = HomeDir + "/templates"
	BinDir       = "/usr/local/bin"

	// relative to the Go embed FS
	SrcConfigDir    = "consul-template/config"
	SrcTemplatesDir = "consul-template/templates"
)

type Flower struct {
	Desc           florist.Description
	Log            hclog.Logger
	FilesFS        fs.FS
	Version        string
	Hash           string
	Configurations []string
	Templates      []string
}

func (fl *Flower) Description() florist.Description {
	if fl.Desc.Name == "" {
		return florist.Description{
			Name: "consultemplate",
			Long: "install the consul-template tool as a service",
		}
	}
	return fl.Desc
}

func (fl *Flower) SetLogger(log hclog.Logger) {
	fl.Log = log
}

func (fl *Flower) Install() error {
	log := fl.Log.Named("flower.consultemplate")
	log.Info("begin")
	defer log.Info("end")

	root, err := user.Current()
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("Add system user 'consul-template'")
	userConsulTemplate, err := florist.UserSystemAdd("consul-template", HomeDir)
	if err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("Create consul-template configuration dir", "dir", ConfigDir)
	if err := florist.Mkdir(ConfigDir, userConsulTemplate, 0755); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("Create consul-template templates dir", "dir", TemplatesDir)
	if err := florist.Mkdir(TemplatesDir, userConsulTemplate, 0755); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	// FIXME SECURITY TODO
	//
	// For the time being we run consul-template as root :-(
	//
	// We should add a dedicated user, consul-template, and then give it, via Linux CAP,
	// "only" the capabilities to write files, doing chown and sending signals to other
	// processes.
	// Actually writing files is so powerful that probably this protection would be
	// useless? :-(
	//
	userConsulTemplate = root

	if err := installConsulTemplateExe(log, fl.Version, fl.Hash, root); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	baseCfg := "aa-base-config.hcl"
	configs := make([]string, 0, len(fl.Configurations)+1)
	configs = append(configs, baseCfg)
	configs = append(configs, fl.Configurations...)

	for _, cfg := range configs {
		src := path.Join(SrcConfigDir, cfg)
		dst := path.Join(ConfigDir, cfg)
		log.Info("Install consul-template configuration file", "dst", dst)
		if err := florist.CopyFromFs(fl.FilesFS, src, dst, 0640, userConsulTemplate); err != nil {
			return fmt.Errorf("%s: %s", log.Name(), err)
		}
	}

	for _, tmpl := range fl.Templates {
		src := path.Join(SrcTemplatesDir, tmpl)
		dst := path.Join(TemplatesDir, tmpl)
		log.Info("Install consul-template template file", "dst", dst)
		if err := florist.CopyFromFs(fl.FilesFS, src, dst, 0640, userConsulTemplate); err != nil {
			return fmt.Errorf("%s: %s", log.Name(), err)
		}
	}

	unit := "consul-template.service"
	src := path.Join("consul-template/service", unit)
	dst := path.Join("/etc/systemd/system/", unit)
	log.Info("Install consul-template systemd unit file", "dst", dst)
	if err := florist.CopyFromFs(fl.FilesFS, src, dst, 0644, root); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	log.Info("Enable service to start at boot", "unit", unit)
	if err := systemd.Enable(unit); err != nil {
		return fmt.Errorf("%s: %s", log.Name(), err)
	}

	// We do not start the service at Packer time because it is not needed.

	return nil
}

func installConsulTemplateExe(log hclog.Logger, version string, hash string, root *user.User) error {
	log.Info("Download consul-template package")
	url := fmt.Sprintf("https://releases.hashicorp.com/consul-template/%s/consul-template_%s_linux_amd64.zip", version, version)
	client := &http.Client{Timeout: 30 * time.Second}
	zipPath, err := florist.NetFetch(client, url, florist.SHA256, hash, florist.WorkDir)
	if err != nil {
		return err
	}

	extracted := path.Join(florist.WorkDir, "consul-template")
	log.Info("Unzipping consul-template package", "dst", extracted)
	if err := florist.UnzipOne(zipPath, "consul-template", extracted); err != nil {
		return err
	}

	exe := path.Join(BinDir, "consul-template")
	log.Info("Install consul-template", "dst", exe)
	if err := florist.Copy(extracted, exe, 0755, root); err != nil {
		return err
	}

	return nil
}
