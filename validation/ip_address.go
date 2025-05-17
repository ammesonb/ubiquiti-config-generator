package validation

import (
	"net"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
)

// IsValidAddress returns true if the address is a valid IP address
func IsValidAddress(address string) bool {
	return net.ParseIP(address) != nil
}

var (
	errInvalidAddress = "invalid IP4 address"
	errInvalidCIDR    = "invalid CIDR"
)

// IsAddressInSubnet returns true if the provided address is in the given CIDR
func IsAddressInSubnet(address string, cidr string) (bool, error) {
	ip := net.ParseIP(address)
	if ip == nil {
		return false, errors.ErrWithCtx(errInvalidAddress, address)
	}

	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, errors.ErrWithCtxParent(errInvalidCIDR, cidr, err)
	}

	return network.Contains(ip), nil
}
