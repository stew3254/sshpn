package main

import (
	"github.com/alecthomas/kong"
)

func main() {
	var cli CLI
	ctx := kong.Parse(&cli, kong.Name("sshpn"))
	err := ctx.Run(&Globals{Version: "0.0.1"})
	ctx.FatalIfErrorf(err)
}
