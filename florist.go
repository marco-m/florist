package florist

import (
	"fmt"
	"os/user"
)

const WorkDir = "/tmp/flower.work"

func Init() (*user.User, error) {
	log := Log.Named("Init")

	userSelf, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("florist.Init: %s", err)
	}
	if err := Mkdir(WorkDir, userSelf, 0755); err != nil {
		return userSelf, fmt.Errorf("florist.Init: %s", err)
	}
	log.Info("success")
	return userSelf, nil
}
