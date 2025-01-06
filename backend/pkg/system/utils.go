package system

import (
	"net"
	"os"
)

func getHostname() string {
	hn, err := os.Hostname()
	if err != nil {
		return ""
	}

	return hn
}

func getIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ips = append(ips, addr.String())
		}
	}

	return ips
}
