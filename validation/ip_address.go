package validation

import (
	"fmt"
	"net"
)

// IsValidAddress returns true if the address is a valid IP address
func IsValidAddress(address string) bool {
	return net.ParseIP(address) != nil
}

// IsAddressInSubnet returns true if the provided address is in the given CIDR
func IsAddressInSubnet(address string, cidr string) (bool, error) {
	ip := net.ParseIP(address)
	if ip == nil {
		return false, fmt.Errorf("Invalid IP4 address '%s'", address)
	}

	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, fmt.Errorf("invalid CIDR '%s': %v", cidr, err)
	}

	return network.Contains(ip), nil
}
