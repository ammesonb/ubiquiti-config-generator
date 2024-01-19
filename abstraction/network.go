package abstraction

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/ammesonb/ubiquiti-config-generator/validation"
)

var errReadNetworkDir = "failed to read networks dir"

func LoadNetworks(networksPath string) ([]Network, []error) {
	var networks []Network

	entries, err := os.ReadDir(networksPath)
	if err != nil {
		return nil, []error{fmt.Errorf("%s %s: %v", errReadNetworkDir, networksPath, err)}
	}

	var errors []error
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		network, errs := loadNetwork(path.Join(networksPath, entry.Name()))
		errors = append(errors, errs...)
		if len(errs) == 0 {
			networks = append(networks, *network)
		}
	}

	return networks, []error{}
}

func loadNetwork(networkPath string) (*Network, []error) {
	config, err := os.ReadFile(path.Join(networkPath, "config.yaml"))
	if err != nil {
		return nil, []error{fmt.Errorf("failed to read network %s/config.yaml: %v", networkPath, err)}
	}
	var network Network

	if err = yaml.Unmarshal(config, &network); err != nil {
		return nil, []error{fmt.Errorf("failed to parse network config at %s: %v", networkPath, err)}
	}

	if network.Interface != nil {
		setupFirewallRuleCounters(network)
	}

	if errs := loadHosts(&network, networkPath); len(errs) > 0 {
		return nil, errs
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

var (
	errFailedReadHostDirectory = "failed to read hosts directory for network"
	errFailedReadHost          = "failed to read host"
	errFailedParseHost         = "failed to parse host"
	errCheckHostSubnet         = "failed checking host subnet"
)

func loadHosts(network *Network, networkPath string) []error {
	// if the hosts directory does not exist, then simply skip loading
	if _, err := os.Stat(path.Join(networkPath, "hosts")); os.IsNotExist(err) {
		return nil
	}

	hostFiles, err := os.ReadDir(path.Join(networkPath, "hosts"))
	if err != nil {
		return []error{fmt.Errorf("%s: %v", errFailedReadHostDirectory, err)}
	}

	errors := make([]error, 0)
	for _, hostFile := range hostFiles {
		if hostFile.IsDir() {
			continue
		}

		hostPath := path.Join(networkPath, "hosts", hostFile.Name())
		hostYAML, err := os.ReadFile(hostPath)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s %s: %v", errFailedReadHost, hostPath, err))
			continue
		}

		nameExtract := regexp.MustCompile(`^(.*)\.ya?ml$`)
		host := Host{
			Name: nameExtract.FindStringSubmatch(hostFile.Name())[1],
		}

		if err = yaml.Unmarshal(hostYAML, &host); err != nil {
			errors = append(errors, fmt.Errorf("%s %s: %v", errFailedParseHost, hostPath, err))
			continue
		}

		for _, subnet := range network.Subnets {
			inSubnet, err := validation.IsAddressInSubnet(host.Address, subnet.CIDR)
			if err != nil {
				errors = append(errors, fmt.Errorf("%s %s: %v", errCheckHostSubnet, hostPath, err))
			} else if inSubnet {
				subnet.Hosts = append(subnet.Hosts, &host)
				break
			}
		}
	}

	return errors
}
