package main

import (
	"github.com/alecthomas/kong"
	"log"
)

var VERSION VersionFlag = "0.0.1"

func main() {
	var cli CLI
	ctx := kong.Parse(
		&cli,
		kong.Name("sshpn"),
		kong.Description("A program to create transparent tunnels using ssh"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{"version": string(VERSION)},
	)
	err := ctx.Run(&Globals{Version: VERSION})
	ctx.FatalIfErrorf(err)
	log.Println(cli.Tun.State.Args)
	log.Println(cli.Tap.State.Args)
	log.Println(cli.Proxy.State.Args)
}
