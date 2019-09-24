package ws

import (
	"encoding/binary"
	"net"
	"strings"
)

func GetLocalhostIp() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return ""
	}

	for _, address := range addrs {

		// 检查ip地址判断是否回环地址
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	return ""
}

func Ip2long(ipstr string) uint32 {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return 0
	}
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}

func IsLocalIP(ip string) bool {
	return IsLocalNet(net.ParseIP(ip))
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

func IsLocalNet(ip net.IP) bool {

	for _, network := range localNetworks {
		if strings.Contains(network, ip.String()) {
			return true
		}
	}

	return ip.IsLoopback()
}
