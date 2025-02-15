package main

import (
	"fmt"
	"os"
	"os/signal"

	zkevmbridgeservice "github.com/fiwallets/zkevm-bridge-service"
	"github.com/fiwallets/zkevm-bridge-service/autoclaimservice/autoclaim"
	"github.com/fiwallets/zkevm-bridge-service/autoclaimservice/blockchainmanager"
	"github.com/fiwallets/zkevm-bridge-service/autoclaimservice/config"
	"github.com/fiwallets/zkevm-bridge-service/log"
	"github.com/urfave/cli/v2"
)

const (
	flagCfg     = "cfg"
	flagNetwork = "network"
)

const (
	// App name
	appName = "zkevm-bridge"
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Version = zkevmbridgeservice.Version
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     flagCfg,
			Aliases:  []string{"c"},
			Usage:    "Configuration `FILE`",
			Required: false,
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:    "version",
			Aliases: []string{},
			Usage:   "Application version and build",
			Action:  version,
		},
		{
			Name:    "run",
			Aliases: []string{},
			Usage:   "Run the zkevm bridge",
			Action:  run,
			Flags:   flags,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}
}

func version(*cli.Context) error {
	zkevmbridgeservice.PrintVersion(os.Stdout)
	return nil
}

func run(ctx *cli.Context) error {
	configFilePath := ctx.String(flagCfg)
	c, err := config.Load(configFilePath)
	if err != nil {
		return err
	}
	log.Init(c.Log)
	bm, err := blockchainmanager.NewClient(ctx.Context, &c.BlockchainManager)
	if err != nil {
		return err
	}
	ac, err := autoclaim.NewAutoClaim(ctx.Context, &c.AutoClaim, bm)
	if err != nil {
		return err
	}
	go ac.Start()

	// Wait for an in interrupt.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	return nil
}
