package main

import (
	"fmt"
	"github.com/alecthomas/kong"
)

type Globals struct {
	Config string `short:"c" default:"/etc/sshpn.conf" type:"existingfile" help:"Location of client config file"`
	// LogLevel string      `short:"l" help:"Set the logging level (debug|info|warn|error|fatal)" default:"info"`
	Version VersionFlag `name:"version" help:"Print version information and quit"`
	Verbose bool        `short:"v" help:"Increase verbosity"`
}

type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

type TunCmd struct {
	Name           string   `short:"n" default:"tun0" help:"The name of the tun device on each system"`
	Laddr          string   `short:"l" default:"10.32.32.2" help:"The local address used in the tunnel"`
	Raddr          string   `short:"r" default:"10.32.32.1" help:"The remote address for the tunnel"`
	All            bool     `short:"a" xor:"subnets" help:"All traffic will be routed through this interface"`
	ExcludeSubnets []string `short:"e" help:"List of subnets to route locally if using the all option"`
	Subnets        []string `short:"s" xor:"subnets" help:"List of subnets to add as routes through the tun"`
}

func (t *TunCmd) Run(globals *Globals) error {
	return nil
}

type TapCmd struct {
	Name string `short:"n" default:"tap0" help:"The name of the tap device on each system"`
	All  bool   `short:"a" help:"All traffic will be routed through this interface"`
	DHCP bool   `short:"d" help:"Once the tap devices are up, dhcp will be used to get an ip on the client"`
}

func (t *TapCmd) Run(globals *Globals) error {
	return nil
}

type ProxyCmd struct {
	Host           string   `short:"h" default:"localhost" help:"The host of the proxy server"`
	Port           uint16   `short:"p" default:"1080" help:"The host of the proxy server"`
	All            bool     `short:"a" xor:"all" help:"All tcp traffic will be routed through this interface and the rest dropped"`
	UDP            bool     `short:"u" help:"The remote supports udp, so all udp will be forwarded as well (not supported by ssh -D)"`
	ExcludeSubnets []string `short:"e" xor:"subnets" help:"List of subnets to route locally if using the all option"`
	Subnets        []string `short:"s" xor:"all,subnets" help:"List of subnets to add as routes through the proxy"`
}

func (p *ProxyCmd) Run(globals *Globals) error {
	return nil
}

type CLI struct {
	Globals
	Tun   TunCmd   `cmd:"" help:"Used to configure a L3 vpn"`
	Tap   TapCmd   `cmd:"" help:"Used to configure a L2 vpn"`
	Proxy ProxyCmd `cmd:"" help:"Used to create a transparent proxy"`
}
