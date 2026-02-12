package conf

import (
	"fmt"
	"net"
)

// GetDefaultInterface attempts to determine the primary network interface
// by dialing an external address (Google DNS) and seeing which local IP is used.
func GetDefaultInterface() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("failed to dial external address: %v", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to list interfaces: %v", err)
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip.Equal(localAddr.IP) {
				return iface.Name, nil
			}
		}
	}

	return "", fmt.Errorf("no interface found matching local IP %s", localAddr.IP)
}
