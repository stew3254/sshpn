package main

import (
	"github.com/alecthomas/kong"
	"strings"
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
	cmd := ctx.Command()

	// Handle running code for the tunnels
	if strings.HasPrefix(cmd, "tun") {
		if cli.Tun.State.State == "start" {
			err = startTun(cli.Tun, cli.Globals)
		} else {
			err = stopTun(cli.Tun, cli.Globals)
		}
	} else if strings.HasPrefix(cmd, "tap") {
		if cli.Tap.State.State == "start" {
			err = startTap(cli.Tap, cli.Globals)
		} else {
			err = stopTap(cli.Tap, cli.Globals)
		}
	} else {
		if cli.Proxy.State.State == "start" {
			err = startProxy(cli.Proxy, cli.Globals)
		} else {
			err = stopProxy(cli.Proxy, cli.Globals)
		}
	}
	ctx.FatalIfErrorf(err)
}
