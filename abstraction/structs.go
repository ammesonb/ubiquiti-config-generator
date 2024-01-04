package abstraction

// PortGroup contains a list of ports for a firewall group
type PortGroup struct {
	Name        string
	Description string `yaml:"description"`
	Ports       []int  `yaml:"ports"`
}

// Network represents a set of subnets with hosts, and can optionally specify interface attributes
type Network struct {
	Name      string
	Interface struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Address     string `yaml:"address"`
		Vif         int32  `yaml:"vif"`
		Speed       string `yaml:"speed"`
		Duplex      string `yaml:"duplex"`

		InboundFirewall  string `yaml:"inbound-firewall"`
		OutboundFirewall string `yaml:"outbound-firewall"`
		LocalFirewall    string `yaml:"local-firewall"`

		Extra map[string]any `yaml:"extra"`
	} `yaml:"interface"`

	Authoritative string   `yaml:"authoritative"`
	Subnets       []Subnet `yaml:"subnets"`
}

// Subnet represents a DHCP-served range of addresses, with some static reservations
type Subnet struct {
	CIDR          string   `yaml:"cidr"`
	DefaultRouter string   `yaml:"default-router"`
	DomainName    string   `yaml:"domain-name"`
	DNSServers    []string `yaml:"dns-servers"`
	DHCPLease     int32    `yaml:"dhcp-lease"`
	DHCPStart     string   `yaml:"dhcp-start"`
	DHCPEnd       string   `yaml:"dhcp-end"`

	Extra map[string]any `yaml:"extra"`

	// These will be determined from other files in the directory using the CIDR
	Hosts []Host
}

// Host contains details about a specific host and firewall connections to allow or deny
type Host struct {
	Address       string   `yaml:"address"`
	MAC           string   `yaml:"mac"`
	AddressGroups []string `yaml:"address-groups"`
	ForwardPorts  []int32  `yaml:"forward-ports"`
	// TODO: hairpin ports?
	Connections []FirewallConnection `yaml:"connections"`
}

// FirewallConnection contains the attributes required to allow or block connections to this host on certain addresses or ports
type FirewallConnection struct {
	Description string `yaml:"description"`
	Allow       bool   `yaml:"allow"`
	Protocol    string `yaml:"protocol"`

	Source      ConnectionDetail `yaml:"source"`
	Destination ConnectionDetail `yaml:"destination"`
}

// ConnectionDetail provides an address and/or port (or group) to allow or block a connection to or from
type ConnectionDetail struct {
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
}
