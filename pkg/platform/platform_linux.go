package platform

import (
	"fmt"
	"os/exec"
	"strings"
)

func collectInfo() (Info, error) {
	info := Info{}

	id, err := exec.Command("lsb_release", "--id", "--short").Output()
	if err != nil {
		return Info{}, fmt.Errorf("CollectInfo: %s", err)
	}
	info.Id = strings.ToLower(strings.TrimSpace(string(id)))

	codename, err := exec.Command("lsb_release", "-cs").Output()
	if err != nil {
		return Info{}, fmt.Errorf("CollectInfo: %s", err)
	}
	info.Codename = strings.ToLower(strings.TrimSpace(string(codename)))

	return info, nil
}
