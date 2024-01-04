package abstraction

import (
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v3"

	"github.com/ammesonb/ubiquiti-config-generator/validation"
)

func LoadNetworks(networksPath string) ([]Network, error) {
	var networks []Network

	entries, err := os.ReadDir(networksPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read networks dir %s: %v", networksPath, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		network, err := loadNetwork(path.Join(networksPath, entry.Name()))
		if err != nil {
			return nil, err
		}

		networks = append(networks, *network)
	}

	return networks, nil
}

func loadNetwork(networkPath string) (*Network, error) {
	config, err := os.ReadFile(path.Join(networkPath, "config.yaml"))
	var network Network

	if err = yaml.Unmarshal(config, &network); err != nil {
		return nil, fmt.Errorf("failed to parse network config at %s: %v", networkPath, err)
	}

	return &network, nil
}

func loadHosts(network *Network, networkPath string) error {
	// if the hosts directory does not exist, then simply skip loading
	if _, err := os.Stat(path.Join(networkPath, "hosts")); os.IsNotExist(err) {
		return nil
	}

	hostFiles, err := os.ReadDir(path.Join(networkPath, "hosts"))
	if err != nil {
		return fmt.Errorf("failed to read hosts directory for network: %v", err)
	}

	for _, hostFile := range hostFiles {
		if hostFile.IsDir() {
			continue
		}

		hostPath := path.Join(networkPath, "hosts", hostFile.Name())
		hostYAML, err := os.ReadFile(hostPath)
		if err != nil {
			return fmt.Errorf("failed to read host %s: %v", hostPath, err)
		}

		var host Host
		if err = yaml.Unmarshal(hostYAML, &host); err != nil {
			return fmt.Errorf("failed to parse host %s: %v", hostPath, err)
		}

		for _, subnet := range network.Subnets {
			inSubnet, err := validation.IsAddressInSubnet(host.Address, subnet.CIDR)
			if err != nil {
				return fmt.Errorf("failed checking host %s subnet: %v", hostPath, err)
			} else if inSubnet {
				subnet.Hosts = append(subnet.Hosts, host)
				break
			}
		}
	}

	return nil
}
