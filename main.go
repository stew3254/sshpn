package main

import (
	"github.com/alecthomas/kong"
)

var VERSION VersionFlag = "0.0.1"

func main() {
	var cli CLI
	ctx := kong.Parse(&cli, kong.Name("sshpn"))
	err := ctx.Run(&Globals{Version: VERSION})
	ctx.FatalIfErrorf(err)
}
