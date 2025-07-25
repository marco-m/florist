//go:build linux

package florist

import (
	"fmt"
	"net"

	"github.com/cakturk/go-netstat/netstat"
	"github.com/marco-m/florist/pkg/sets"
)

func ListeningSockets() (*sets.Set[string], error) {
	ss := sets.New[string](0)

	// UDPv4
	socks, err := netstat.UDPSocks(func(s *netstat.SockTabEntry) bool {
		return true
	})
	if err != nil {
		return nil, err
	}
	for _, s := range socks {
		ss.Add(fmt.Sprintf("udp4 %s", s.LocalAddr.String()))
	}

	// UDPv6
	socks, err = netstat.UDP6Socks(func(s *netstat.SockTabEntry) bool {
		return true
	})
	if err != nil {
		return nil, err
	}
	for _, s := range socks {
		ss.Add(fmt.Sprintf("udp6 %s", s.LocalAddr.String()))
	}

	// TCPv4
	socks, err = netstat.TCPSocks(func(s *netstat.SockTabEntry) bool {
		return s.State == netstat.Listen
	})
	if err != nil {
		return nil, err
	}
	for _, s := range socks {
		ss.Add(fmt.Sprintf("tcp4 %s", s.LocalAddr.String()))
	}

	// TCPv6
	socks, err = netstat.TCP6Socks(func(s *netstat.SockTabEntry) bool {
		return s.State == netstat.Listen
	})
	if err != nil {
		return nil, err
	}
	for _, s := range socks {
		ss.Add(fmt.Sprintf("tcp6 %s", s.LocalAddr.String()))
	}

	return ss, nil
}

func PrivateIPs() ([]string, error) {
	var ips []string

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				return nil, err
			}
			if ip.IsGlobalUnicast() && !ip.IsPrivate() {
				continue
			}
			ips = append(ips, ip.String())
		}
	}

	return ips, nil
}

func PublicIPs() ([]string, error) {
	var ips []string

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				return nil, err
			}
			if !ip.IsGlobalUnicast() || ip.IsPrivate() {
				continue
			}
			ips = append(ips, ip.String())
		}
	}

	return ips, nil
}

// PrivateIP returns the first matching IP that belongs to the given network CIDR.
func PrivateIP(cidr string) (string, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("network CIDR: %s: error: %s", cidr, err)
	}

	ips, err := PrivateIPs()
	if err != nil {
		return "", err
	}

	for _, addr := range ips {
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}

		if network.Contains(ip) {
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("none of the private IPs belongs to %s network", cidr)
}
