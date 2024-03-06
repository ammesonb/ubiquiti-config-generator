package validation

import (
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidAddress(t *testing.T) {
	assert.True(t, IsValidAddress("192.168.0.1"))
	assert.True(t, IsValidAddress("127.0.0.1"))
	assert.False(t, IsValidAddress("abcdef"))
	assert.False(t, IsValidAddress("127001"))
}

func TestIsAddressInSubnet(t *testing.T) {
	cidr := "foo"
	addr := "abcdef"
	inSubnet, err := IsAddressInSubnet(addr, "foo")
	assert.Falsef(t, inSubnet, "Invalid address should not be in subnet")
	assert.ErrorIs(t, err, utils.ErrWithCtx(errInvalidAddress, addr))

	inSubnet, err = IsAddressInSubnet("1.1.1.1", cidr)
	assert.Falsef(t, inSubnet, "Invalid CIDR should not be in subnet")
	assert.ErrorIs(t, err, utils.ErrWithCtx(errInvalidCIDR, cidr))

	cidr = "192.168.0.0/33"
	inSubnet, err = IsAddressInSubnet("1.1.1.1", cidr)
	assert.Falsef(t, inSubnet, "Invalid CIDR should not be in subnet")
	assert.ErrorIs(t, err, utils.ErrWithCtx(errInvalidCIDR, cidr))

	inSubnet, err = IsAddressInSubnet("1.1.1.1", "192.168.0.0/32")
	assert.Falsef(t, inSubnet, "Address should not be in subnet")
	assert.NoError(t, err)

	inSubnet, err = IsAddressInSubnet("192.168.0.1", "192.168.0.0/24")
	assert.Truef(t, inSubnet, "Address should be in subnet")
	assert.NoError(t, err)

	inSubnet, err = IsAddressInSubnet("192.168.0.255", "192.168.0.0/24")
	assert.Truef(t, inSubnet, "Address should be in subnet")
	assert.NoError(t, err)
}
