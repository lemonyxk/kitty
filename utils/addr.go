package utils

import (
	"encoding/binary"
	"net"
	"strings"
)

var defaultIP = ""

type addr int

const Addr addr = iota

func (a addr) GetLocalhostIp() string {

	if defaultIP != "" {
		return defaultIP
	}

	addresses, err := net.InterfaceAddrs()

	if err != nil {
		defaultIP = "127.0.0.1"
		return defaultIP
	}

	for i := 0; i < len(addresses); i++ {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := addresses[i].(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				defaultIP = ipNet.IP.String()
				return defaultIP
			}
		}
	}

	defaultIP = "127.0.0.1"
	return defaultIP
}

func (a addr) Ip2long(ipStr string) uint32 {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0
	}
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}

func (a addr) IsLocalIP(ip string) bool {
	return a.IsLocalNet(net.ParseIP(ip))
}

var localNetworks = []string{
	"10.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"172.17.0.0/12",
	"172.18.0.0/12",
	"172.19.0.0/12",
	"172.20.0.0/12",
	"172.21.0.0/12",
	"172.22.0.0/12",
	"172.23.0.0/12",
	"172.24.0.0/12",
	"172.25.0.0/12",
	"172.26.0.0/12",
	"172.27.0.0/12",
	"172.28.0.0/12",
	"172.29.0.0/12",
	"172.30.0.0/12",
	"172.31.0.0/12",
	"192.168.0.0/16",
}

func (a addr) IsLocalNet(ip net.IP) bool {

	for i := 0; i < len(localNetworks); i++ {
		if strings.Contains(localNetworks[i], ip.String()) {
			return true
		}
	}

	return ip.IsLoopback()
}
