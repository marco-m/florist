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

func (cmd *listCmd) Run(app App) error {
	for _, k := range app.prov.ordered {
		v := app.prov.flowers[k]
		if err := v.Init(); err != nil {
			return err
		}
		fmt.Printf("%s -- %s\n", v, v.Description())
		for _, fi := range v.Embedded() {
			fmt.Printf("  %s\n", fi)
		}
	}
	return nil
}
