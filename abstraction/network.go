package abstraction

import (
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"os"
	"path"
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/ammesonb/ubiquiti-config-generator/validation"
)

var (
	errReadNetworks     = "failed to read networks in directory %s"
	errReadNetworkConf  = "failed to read network config in %s"
	errParseNetworkConf = "failed to parse network config in %s"
	errReadHostDir      = "failed to read hosts directory for network: %s"
	errReadHost         = "failed to read host: %s"
	errParseHost        = "failed to parse host: %s"
	errCheckHostSubnet  = "failed checking host subnet: %s"
)

func LoadNetworks(networksPath string) ([]Network, []error) {
	var networks []Network

	entries, err := os.ReadDir(networksPath)
	if err != nil {
		return nil, []error{utils.ErrWithCtxParent(errReadNetworks, networksPath, err)}
	}

	var errors []error
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		network, errs := loadNetwork(path.Join(networksPath, entry.Name()))
		errors = append(errors, errs...)
		if len(errs) == 0 && network != nil {
			networks = append(networks, *network)
		}
	}

	return networks, errors
}

func loadNetwork(networkPath string) (*Network, []error) {
	config, err := os.ReadFile(path.Join(networkPath, "config.yaml"))
	if err != nil {
		return nil, []error{utils.ErrWithCtxParent(errReadNetworkConf, networkPath, err)}
	}
	var network Network

	if err = yaml.Unmarshal(config, &network); err != nil {
		return nil, []error{utils.ErrWithCtxParent(errParseNetworkConf, networkPath, err)}
	}

	if errs := loadHosts(&network, networkPath); len(errs) > 0 {
		return nil, errs
	}

	if network.Interface != nil {
		setupFirewallRuleCounters(network)
	}

	return &network, []error{}
}

func setupFirewallRuleCounters(network Network) {
	for _, firewall := range []string{
		network.Interface.InboundFirewall,
		network.Interface.OutboundFirewall,
		network.Interface.LocalFirewall,
	} {
		if exists := HasCounter(firewall); !exists {
			MakeCounter(firewall, network.FirewallRuleNumberStart, network.FirewallRuleNumberStep)
		}
	}
}

func loadHosts(network *Network, networkPath string) []error {
	// if the hosts directory does not exist, then simply skip loading
	if _, err := os.Stat(path.Join(networkPath, "hosts")); os.IsNotExist(err) {
		return nil
	}

	hostDir := path.Join(networkPath, "hosts")
	hostFiles, err := os.ReadDir(hostDir)
	if err != nil {
		return []error{utils.ErrWithCtxParent(errReadHostDir, hostDir, err)}
	}

	errors := make([]error, 0)
	for _, hostFile := range hostFiles {
		if !hostFile.Type().IsRegular() {
			continue
		}

		if err = loadHost(
			path.Join(networkPath, "hosts", hostFile.Name()),
			network,
		); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func loadHost(hostPath string, network *Network) error {
	hostYAML, err := os.ReadFile(hostPath)
	if err != nil {
		return utils.ErrWithCtxParent(errReadHost, hostPath, err)
	}

	nameExtract := regexp.MustCompile(`^.*/(.*)\.ya?ml$`)
	matches := nameExtract.FindStringSubmatch(hostPath)
	// If not a YAML file, skip it
	if len(matches) == 0 {
		return nil
	}

	host := Host{
		Name: matches[1],
	}

	if err = yaml.Unmarshal(hostYAML, &host); err != nil {
		return utils.ErrWithCtxParent(errParseHost, hostPath, err)
	}

	for _, subnet := range network.Subnets {
		inSubnet, err := validation.IsAddressInSubnet(host.Address, subnet.CIDR)
		if err != nil {
			return utils.ErrWithCtxParent(errCheckHostSubnet, hostPath, err)
		} else if inSubnet {
			subnet.Hosts = append(subnet.Hosts, &host)
		}
	}

	return nil
}
