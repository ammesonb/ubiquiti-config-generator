package abstraction

import (
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMissingDirectory(t *testing.T) {
	network, errs := LoadNetworks("/nonexistent")
	assert.Nil(t, network, "No network returned if path does not exist")
	assert.Len(t, errs, 1, "Only one error returned on missing networks directory")
	assert.ErrorIs(t, errs[0], utils.ErrWithCtx(errReadNetworks, "/nonexistent"))
}

func TestLoadNetworks(t *testing.T) {
	networks, errs := LoadNetworks("test-files/networks")
	assert.NotNil(t, errs, "Errors should not be nil")
	assert.NotEmpty(t, errs, "Some errors should be returned")
	assert.Len(t, errs, 5, "Four errors returned")
	assert.ErrorIs(
		t,
		errs[0],
		utils.ErrWithCtx(errParseHost, "test-files/networks/failing-hosts/hosts/incorrect-syntax.yaml"),
	)
	assert.ErrorIs(
		t,
		errs[1],
		utils.ErrWithCtx(errCheckHostSubnet, "test-files/networks/failing-hosts/hosts/invalid-address.yaml"),
	)
	assert.ErrorIs(
		t,
		errs[2],
		utils.ErrWithCtx(errReadNetworkConf, "test-files/networks/hosts-as-file"),
	)
	assert.ErrorIs(t, errs[3], utils.ErrWithCtx(errParseNetworkConf, "test-files/networks/invalid-config"))
	assert.ErrorIs(t, errs[4], utils.ErrWithCtx(errReadNetworkConf, "test-files/networks/missing-config"))

	assert.NotNil(t, networks, "Networks should be returned")
	assert.Len(t, networks, 1, "One network should be valid")
	assert.Len(t, networks[0].Subnets, 2, "Two subnets parsed")
	assert.Len(t, networks[0].Subnets[0].Hosts, 1, "One host in first subnet")
	assert.Len(t, networks[0].Subnets[1].Hosts, 1, "One host in second subnet")
}
func TestSetupFirewallCounters(t *testing.T) {
	resetCounters()

	assert.False(t, HasCounter("inbound"), "inbound firewall does not exist yet")
	assert.False(t, HasCounter("outbound"), "outbound firewall does not exist yet")
	assert.False(t, HasCounter("local"), "local firewall does not exist yet")

	setupFirewallRuleCounters(Network{Interface: &Interface{
		InboundFirewall:  "inbound",
		OutboundFirewall: "outbound",
		LocalFirewall:    "local",
	}})

	assert.True(t, HasCounter("inbound"), "inbound firewall created")
	assert.True(t, HasCounter("outbound"), "outbound firewall created")
	assert.True(t, HasCounter("local"), "local firewall created")
	assert.False(t, HasCounter("other"), "other firewall does not exist")
}

func TestLoadNetworkFailures(t *testing.T) {
	network, errs := loadNetwork("./test-files/networks/missing-config")
	assert.Nil(t, network, "No network returned if config is missing")
	assert.ErrorIs(
		t,
		errs[0],
		utils.ErrWithCtx(
			errReadNetworkConf,
			"./test-files/networks/missing-config",
		),
	)

	network, errs = loadNetwork("./test-files/networks/invalid-config")
	assert.Nil(t, network, "No network returned if config is invalid")
	assert.ErrorIs(t, errs[0], utils.ErrWithCtx(errParseNetworkConf, "./test-files/networks/invalid-config"))

	network, errs = loadNetwork("./test-files/networks/failing-hosts")
	assert.NotNil(t, errs, "Errors for hosts should be returned")
	assert.Len(t, errs, 2, "Two hosts have errors")
	assert.ErrorIs(
		t,
		errs[0],
		utils.ErrWithCtx(errParseHost, "test-files/networks/failing-hosts/hosts/incorrect-syntax.yaml"),
		"First error is for invalid host YAML",
	)
	assert.ErrorIs(
		t,
		errs[1],
		utils.ErrWithCtx(errCheckHostSubnet, "test-files/networks/failing-hosts/hosts/invalid-address.yaml"),
		"Second error is for host address check",
	)

	assert.Nil(t, network, "Network should be nil if hosts have errors")
	assert.False(t, HasCounter("eth1-in"), "Inbound firewall counter not set up")
	assert.False(t, HasCounter("eth1-out"), "Outbound firewall counter not set up")
	assert.False(t, HasCounter("eth1-local"), "Local firewall counter not set up")
}

func TestLoadNetworkSuccess(t *testing.T) {
	network, errs := loadNetwork("./test-files/networks/successful-hosts")
	assert.Empty(t, errs, "No errors for hosts should be returned")

	assert.NotNil(t, network, "Network should not be nil")
	assert.True(t, HasCounter("eth1-in"), "Inbound firewall counter set up")
	assert.True(t, HasCounter("eth1-out"), "Outbound firewall counter set up")
	assert.True(t, HasCounter("eth1-local"), "Local firewall counter set up")
	assert.Equal(t, 10, GetCounter("eth1-in").step, "Step configured for firewall")
	assert.Equal(t, 1000, GetCounter("eth1-in").number, "Starting rule number configured for firewall")

	assert.Len(t, network.Subnets[0].Hosts, 1, "One host loaded for first subnet")
	assert.Len(t, network.Subnets[1].Hosts, 1, "One hosts loaded for second subnet")
}

func TestLoadHost(t *testing.T) {
	network := &Network{}

	err := loadHost("/nonexistent", network)
	assert.ErrorIs(
		t,
		err,
		utils.ErrWithCtx(errReadHost, "/nonexistent"),
		"Should fail to read nonexistent host",
	)

	hostDirectory := "test-files/networks/failing-hosts/hosts/"
	err = loadHost(hostDirectory+"incorrect-syntax.yaml", network)
	assert.ErrorIs(
		t,
		err,
		utils.ErrWithCtx(errParseHost, "test-files/networks/failing-hosts/hosts/incorrect-syntax.yaml"),
	)
	assert.ErrorContains(t, err, hostDirectory+"incorrect-syntax.yaml")

	network.Subnets = []*Subnet{
		{
			CIDR:  "10.0.1.0/24",
			Hosts: make([]*Host, 0),
		},
		{
			CIDR:  "10.0.0.0/24",
			Hosts: make([]*Host, 0),
		},
	}

	err = loadHost(hostDirectory+"invalid-address.yaml", network)
	assert.ErrorIs(
		t,
		err,
		utils.ErrWithCtx(errCheckHostSubnet, "test-files/networks/failing-hosts/hosts/invalid-address.yaml"),
	)
	assert.ErrorContains(t, err, hostDirectory+"invalid-address.yaml")

	err = loadHost(hostDirectory+"firewalled-host.yaml", network)
	assert.Nil(t, err, "No error thrown when loading valid host")

	assert.Len(t, network.Subnets[0].Hosts, 0, "No hosts added to first subnet")
	assert.Len(t, network.Subnets[1].Hosts, 1, "Host added to second subnet")
}

func TestLoadHosts(t *testing.T) {
	network := &Network{
		Subnets: []*Subnet{

			{
				CIDR:  "10.16.0.0/24",
				Hosts: make([]*Host, 0),
			},
			{
				CIDR:  "10.0.0.0/24",
				Hosts: make([]*Host, 0),
			},
		},
	}

	errs := loadHosts(network, "/nonexistent")
	assert.Nil(t, errs, "No error for nonexistent host directory")
	assert.Empty(t, network.Subnets[0].Hosts, "No hosts loaded for missing directory")
	assert.Empty(t, network.Subnets[1].Hosts, "No hosts loaded for missing directory")

	errs = loadHosts(network, "./test-files/networks/hosts-as-file")
	assert.NotNil(t, errs, "Error should be returned")
	assert.Len(t, errs, 1, "Returned one error on fail read hosts dir")
	assert.ErrorIs(t, errs[0], utils.ErrWithCtx(errReadHostDir, "test-files/networks/hosts-as-file/hosts"))
	assert.Empty(t, network.Subnets[0].Hosts, "No hosts loaded for invalid hosts directory")
	assert.Empty(t, network.Subnets[1].Hosts, "No hosts loaded for invalid hosts directory")

	errs = loadHosts(network, "./test-files/networks/failing-hosts")
	assert.NotNil(t, errs, "Error should be returned")
	assert.Len(t, network.Subnets[0].Hosts, 1, "One host loaded for first subnet")
	assert.Len(t, network.Subnets[1].Hosts, 1, "One hosts loaded for second subnet")

	assert.Len(t, errs, 2, "Two hosts have errors")
	assert.ErrorIs(
		t,
		errs[0],
		utils.ErrWithCtx(errParseHost, "test-files/networks/failing-hosts/hosts/incorrect-syntax.yaml"),
		"First error is for invalid host YAML",
	)
	assert.ErrorIs(
		t,
		errs[1],
		utils.ErrWithCtx(errCheckHostSubnet, "test-files/networks/failing-hosts/hosts/invalid-address.yaml"),
		"Second error is for host address check",
	)
}
