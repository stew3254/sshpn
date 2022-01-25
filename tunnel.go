package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var deviceCmd = `ip r | grep -E "^default" | grep -oE "dev [a-z0-9]+" | cut -d " " -f 2`
var publicIPCmd = `dig +short myip.opendns.com @resolver1.opendns.com +timeout=3 || curl ifconfig.me || echo Failed`
var routeCmd = `ip r | grep -E '^default' | grep -oE '([0-9]{1,3}[\.]){3}[0-9]{1,3}'`

func iptablesLineGen(laddr string) string {
	return fmt.Sprintf(
		"iptables -t nat -L POSTROUTING --line-number | grep '%s' | grep -oE '^[0-9]+'",
		laddr,
	)
}

// sshInitGen creates an ssh command for starting a tunnel
func sshInitGen(dev, socket, host string) (out string) {
	if _, err := strconv.Atoi(dev); err == nil {
		// Create the start of the command
		out = fmt.Sprintf(
			"ssh -fNTMS '%s' -D %s root@%s",
			socket,
			dev,
			host,
		)
	} else {
		// Set the tunnel type
		var tunType string
		if strings.HasPrefix(dev, "tun") {
			tunType = "point-to-point"
		} else if strings.HasPrefix(dev, "tap") {
			tunType = "ethernet"
		}

		// Create the start of the command
		out = fmt.Sprintf(
			"ssh -fNTMS '%s' -w %s:%s -o Tunnel=%s root@%s",
			socket,
			dev[3:],
			dev[3:],
			tunType,
			host,
		)
	}

	return out
}

// sshStopGen creates an ssh command for stopping a tunnel
func sshStopGen(socket, host string) (out string) {
	// Create the start of the command
	// Exit probably isn't clean, but stop doesn't seem to work and no issues have occurred with exit
	out = fmt.Sprintf(
		"ssh -S '%s' -o exit root@%s",
		socket,
		host,
	)

	return out
}

// sshRunGen creates the string for running commands
func sshRunGen(socket, host, command string) (out string) {
	// Create the command
	out = fmt.Sprintf(
		"ssh -S '%s' root@%s -- '(%s)'",
		socket,
		host,
		command,
	)

	return out
}

func startTun(t TunCmd, globals Globals) error {
	// Start the tunnel and create the devices
	fmt.Println(sshInitGen(t.Name, globals.Socket, t.State.Host))
	// Get the name of the remote's default interface and ip address
	fmt.Println(sshRunGen(globals.Socket, t.State.Host, deviceCmd))
	fmt.Println(sshRunGen(globals.Socket, t.State.Host, publicIPCmd))
	var publicIP string = "1.1.1.1"

	var rdev string = "eth0"
	// Create remote commands to run
	// TODO check if the local and remote ips are in the subnets provided
	// Add the ip to the tunnel
	remoteCmds := fmt.Sprintf("ip a add %s/32 peer %s dev tun0", t.Raddr, t.Laddr)
	// Create the forward rule so we can NAT through the host
	remoteCmds += fmt.Sprintf(
		" && iptables -t nat -A POSTROUTING %s/32 -o %s -j MASQUERADE",
		t.Laddr,
		rdev,
	)
	// Backup the ip forward state
	remoteCmds += " && cp /proc/sys/net/ipv4/ip_forward /tmp/ip_forward"
	// Make sure forward is on
	remoteCmds += " && echo 1 > /proc/sys/net/ipv4/ip_forward"
	// Set the tunel up
	remoteCmds += " && ip l set tun0 up"
	fmt.Println(sshRunGen(globals.Socket, t.State.Host, remoteCmds))

	// Write the public ip to tmp so that we can remember the ip address later
	err := ioutil.WriteFile("/tmp/vpn_ip", []byte(publicIP), 0600)
	if err != nil {
		return err
	}

	// Get the local default route device
	fmt.Println(deviceCmd)
	dev := "wlp0s20f3"

	// Set the ip on the tunnel
	localCmds := fmt.Sprintf("sudo ip a add %s/32 peer %s dev %s", t.Laddr, t.Raddr, t.Name)
	// Set it up
	localCmds += fmt.Sprintf(" && sudo ip l set %s up", t.Name)
	// Make sure we route our traffic to the tunnel correctly on the outbound interface,
	// otherwise packets will drop
	localCmds += fmt.Sprintf(" && sudo ip r add %s via $(%s) dev %s", publicIP, routeCmd, dev)

	if t.All {
		// All traffic should be routed, add global routes
		localCmds += fmt.Sprintf(" && sudo ip r add 0.0.0.0/1 dev %s", t.Name)
		localCmds += fmt.Sprintf(" && sudo ip r add 128.0.0.0/1 dev %s", t.Name)

		// Make sure to add any excluded subnets to still route locally
		// NOTE this assumes that you want to use the default outbound interface, if you don't,
		// you will need to add these routes manually
		for _, subnet := range t.ExcludeSubnets {
			localCmds += fmt.Sprintf(" && sudo ip r add %s dev %s", subnet, dev)
		}
	} else {
		// Make sure to add any subnets to route over the new connection
		for _, subnet := range t.Subnets {
			localCmds += fmt.Sprintf(" && sudo ip r add %s dev %s", subnet, t.Name)
		}
	}

	fmt.Println(localCmds)
	if !globals.Quiet {
		fmt.Println("L3 tunnel created successfully")
	}

	return nil
}

func stopTun(t TunCmd, globals Globals) error {
	// Recall the remote ip
	publicIPBytes, err := ioutil.ReadFile("/tmp/vpn_ip")
	if err != nil {
		return err
	}
	publicIP := strings.TrimSpace(string(publicIPBytes))

	// Get the number of the iptables rule to destroy
	fmt.Println(sshRunGen(globals.Socket, t.State.Host, iptablesLineGen(t.Laddr)))
	nums := strings.Fields("1\n2")
	var remoteCmds string
	if len(nums) > 1 || len(nums[0]) > 0 {
		// Destroy all the rules for this ip (there should only be 1, but just in case)
		for i := len(nums) - 1; i >= 0; i-- {
			remoteCmds += fmt.Sprintf("iptables -t nat -D POSTROUTING %s && ", nums[i])
		}
	}

	// Restore the forward state
	remoteCmds += "cat /tmp/ip_forward > /proc/sys/net/ipv4/ip_forward && rm /tmp/ip_forward"
	fmt.Println(sshRunGen(globals.Socket, t.State.Host, remoteCmds))

	// Close the ssh connection
	fmt.Println(sshStopGen(globals.Socket, t.State.Host))

	// Remove the public ip file
	err = os.Remove("/tmp/vpn_ip")
	if err != nil {
		return err
	}

	// Get the global routing device
	fmt.Println(deviceCmd)
	dev := "wlp0s20f3"

	// Clean up local routing
	localCmds := fmt.Sprintf("sudo ip r del %s dev %s", publicIP, dev)

	// Clean up unnecessary routes
	for _, subnet := range t.ExcludeSubnets {
		localCmds += fmt.Sprintf(" && sudo ip r del %s dev %s", subnet, dev)
	}

	fmt.Println(localCmds)
	if !globals.Quiet {
		fmt.Println("L3 tunnel stopped successfully")
	}

	return nil
}

func startTap(t TapCmd, globals Globals) error {
	fmt.Println(sshInitGen(t.Name, globals.Socket, t.State.Host))
	fmt.Println(sshRunGen(globals.Socket, t.State.Host, "echo hi"))
	return nil
}

func stopTap(t TapCmd, globals Globals) error {
	return nil
}

func startProxy(p ProxyCmd, globals Globals) error {
	fmt.Println(sshInitGen(strconv.Itoa(int(p.Port)), globals.Socket, p.State.Host))
	fmt.Println(sshRunGen(globals.Socket, p.State.Host, "echo hi"))
	return nil
}

func stopProxy(p ProxyCmd, globals Globals) error {
	return nil
}
