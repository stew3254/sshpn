package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/pkg/errors"
	"net"
	"regexp"
)

var tunExp = regexp.MustCompile("tun[0-9]+")
var tapExp = regexp.MustCompile("tap[0-9]+")
var hostExp = regexp.MustCompile(
	"^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*" +
		"([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])$",
)

// Validate CIDR
func validateCIDR(subnets []string) bool {
	if len(subnets) > 0 {
		for _, cidr := range subnets {
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				return false
			}
		}
	}
	return true
}

type Globals struct {
	Version VersionFlag `help:"Print version information and quit"`
	Verbose bool        `short:"v" help:"Increase verbosity"`
	Quiet   bool        `short:"q" help:"Minimize command output"`
	Socket  string      `short:"S" default:"/root/.ssh/socket/%h:%p" help:"Location of master ssh multiplex socket"`
}

type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	// TODO figure out why this is empty
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

type State struct {
	State string `arg:"" help:"Start or stop a tunnel"`
	Host  string `arg:"" passthrough:"" help:"The host you're connecting to via ssh'"`
}

func (s *State) Validate() error {
	if s.State != "start" && s.State != "stop" {
		return errors.New("invalid state, must be 'start' or 'stop'")
	}
	if !hostExp.Match([]byte(s.Host)) && net.ParseIP(s.Host) == nil {
		return errors.New("invalid hostname")
	}

	return nil
}

type TunCmd struct {
	Name           string   `short:"n" default:"tun0" help:"The name of the tun device on each system"`
	Laddr          string   `short:"l" default:"10.32.32.2" help:"The local address used in the tunnel. (Don't change this unless you know what you're doing)"`
	Raddr          string   `short:"r" default:"10.32.32.1" help:"The remote address for the tunnel. (Don't change this unless you know what you're doing)"`
	All            bool     `short:"a" xor:"subnets" help:"All traffic will be routed through this interface"`
	ExcludeSubnets []string `short:"e" help:"List of subnets to route locally if using the all option"`
	Subnets        []string `short:"s" xor:"subnets" help:"List of subnets to add as routes through the tun"`
	State          State    `arg:""`
}

func (t *TunCmd) Validate() error {
	// Validate all inputs so command injection isn't possible
	// Validate name
	if !tunExp.Match([]byte(t.Name)) {
		return errors.New("invalid name, must be of format tun[0-9]+")
	}

	// Validate ip
	if net.ParseIP(t.Laddr) == nil {
		return errors.New(fmt.Sprintf("invalid ip address '%s'", t.Laddr))
	} else if net.ParseIP(t.Raddr) == nil {
		return errors.New(fmt.Sprintf("invalid ip address '%s'", t.Raddr))
	}

	if !validateCIDR(t.ExcludeSubnets) || !validateCIDR(t.Subnets) {
		return errors.New("invalid subnet, must use cidr notation")
	}

	return nil
}

func (t *TunCmd) Run(globals *Globals) error {
	return nil
}

type TapCmd struct {
	Name  string `short:"n" default:"tap0" help:"The name of the tap device on each system"`
	All   bool   `short:"a" help:"All traffic will be routed through this interface"`
	DHCP  bool   `short:"d" help:"Once the tap devices are up, dhcp will be used to get an ip on the client"`
	State State  `arg:""`
}

func (t *TapCmd) Validate() error {
	// Validate all inputs so command injection isn't possible
	// Validate name
	if !tapExp.Match([]byte(t.Name)) {
		return errors.New("invalid name, must be of format tap[0-9]+")
	}

	return nil
}

func (t *TapCmd) Run(globals *Globals) error {
	return nil
}

type ProxyCmd struct {
	Host           string   `short:"h" default:"localhost" help:"The host of the proxy server"`
	Port           uint16   `short:"p" default:"1080" help:"The port of the proxy server"`
	All            bool     `short:"a" xor:"all" help:"All tcp traffic will be routed through this interface and the rest dropped"`
	UDP            bool     `short:"u" help:"The remote supports udp, so all udp will be forwarded as well (not supported by ssh -D)"`
	ExcludeSubnets []string `short:"e" xor:"subnets" help:"List of subnets to route locally if using the all option"`
	Subnets        []string `short:"s" xor:"all,subnets" help:"List of subnets to add as routes through the proxy"`
	State          State    `arg:""`
}

func (p *ProxyCmd) Validate() error {
	if !hostExp.Match([]byte(p.Host)) && net.ParseIP(p.Host) == nil {
		return errors.New("invalid hostname")
	}

	if !validateCIDR(p.ExcludeSubnets) || !validateCIDR(p.Subnets) {
		return errors.New("invalid subnet, must use cidr notation")
	}

	return nil
}

func (p *ProxyCmd) Run(globals *Globals) error {
	return nil
}

type CLI struct {
	Globals
	Tun   TunCmd   `cmd:"" required:"" help:"Used to configure a L3 vpn"`
	Tap   TapCmd   `cmd:"" required:"" help:"Used to configure a L2 vpn"`
	Proxy ProxyCmd `cmd:"" required:"" help:"Used to create a transparent proxy"`
}
