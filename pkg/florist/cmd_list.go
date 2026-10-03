package florist

import (
	"fmt"

	"github.com/marco-m/clim"
)

type listCmd struct{}

func newListCmd(parent *clim.CLI) (*listCmd, error) {
	listCmd := listCmd{}

	_, err := clim.NewSub(parent, "list", "list the flowers and their files")
	if err != nil {
		return nil, err
	}
	return &listCmd, err
}

func (cmd *listCmd) Run(gdn *Garden) error {
	for _, k := range gdn.prov.ordered {
		fl := gdn.prov.flowers[k]
		if err := fl.Init(gdn); err != nil {
			return err
		}
		fmt.Printf("%s -- %s\n", fl, fl.Description())
		for _, fi := range fl.Embedded() {
			fmt.Printf("  %s\n", fi)
		}
	}
	return nil
}
