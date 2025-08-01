package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/marco-m/clim"
)

func main() {
	os.Exit(MainInt())
}

func MainInt() int {
	err := mainErr(os.Args[1:])
	if err == nil {
		return 0
	}
	if errors.Is(err, clim.ErrHelp) {
		fmt.Fprintln(os.Stdout, err)
		return 0
	}
	fmt.Fprintln(os.Stderr, err)
	return 1
}

type Application struct {
	MaxWait      time.Duration
	PollInterval time.Duration
	Address      string
	Verbose      bool
	//
}

func mainErr(args []string) error {
	var app Application
	cli, err := clim.NewTop("wait-for-conn", "waits for a network service to be up", app.run)
	if err != nil {
		return err
	}

	if err := cli.AddFlags(
		&clim.Flag{
			Value: clim.Duration(&app.MaxWait, 15*time.Second),
			Long:  "max-wait", Label: "DURATION", Help: "Max wait time before giving up",
		},
		&clim.Flag{
			Value: clim.Duration(&app.PollInterval, 2*time.Second),
			Long:  "poll-interval", Label: "DURATION", Help: "Poll interval",
		},
		&clim.Flag{
			Value: clim.String(&app.Address, ""),
			Long:  "address", Label: "HOST:PORT", Help: "Address",
			Required: true,
		},
		&clim.Flag{
			Value: clim.Bool(&app.Verbose, false),
			Long:  "verbose", Help: "Be verbose",
		},
	); err != nil {
		return err
	}

	action, err := cli.Parse(args)
	if err != nil {
		return err
	}

	return action(0)
}

func (app *Application) run(uctx int) error {
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), app.MaxWait)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err := net.DialTimeout("tcp", app.Address, app.PollInterval)
			if err != nil {
				if app.Verbose {
					fmt.Println(err)
				}
				break
			}
			elapsed := time.Since(now).Round(time.Second)
			fmt.Println("connected to", app.Address, "after", elapsed)
			return nil
		}
		time.Sleep(app.PollInterval)
	}
}
