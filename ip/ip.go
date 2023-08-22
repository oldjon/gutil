package ip

import "net"

// list of private subnets
var privateMasks, _ = toMasks([]string{
	"127.0.0.0/8",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"fc00::/7",
})

// IsPublicIP returns true if the given IP is not in private ip mask range
func IsPublicIP(ip net.IP) bool {
	return !IsPrivateIP(ip)
}

// IsPrivateIP returns true if the given IP is in private ip mask range
func IsPrivateIP(ip net.IP) bool {
	return ipInMasks(ip, privateMasks)
}

// toMasks converts a list of subnets' string to a list of net.IPNet.
func toMasks(ips []string) (masks []net.IPNet, err error) {
	for _, cidr := range ips {
		var network *net.IPNet
		_, network, err = net.ParseCIDR(cidr)
		if err != nil {
			return
		}
		masks = append(masks, *network)
	}
	return
}

//ipInMasks checks if a net.IP is in a list of net.IPNet
func ipInMasks(ip net.IP, masks []net.IPNet) bool {
	for _, mask := range masks {
		if mask.Contains(ip) {
			return true
		}
	}
	return false
}
