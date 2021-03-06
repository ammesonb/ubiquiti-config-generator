"""
Contains the root configuration node
"""
import ipaddress
from os import path
from typing import List, Tuple

from ubiquiti_config_generator import file_paths, secondary_configs
from ubiquiti_config_generator.nodes import (
    GlobalSettings,
    PortGroup,
    ExternalAddresses,
    NAT,
    Network,
)


class RootNode:
    """
    Represents the root config node, from which everything else is based off of
    """

    # pylint: disable=too-many-arguments
    def __init__(
        self,
        global_settings: GlobalSettings,
        port_groups: List[PortGroup],
        external_addresses: ExternalAddresses,
        networks: List[Network],
        nat: NAT,
    ):
        self.global_settings = global_settings
        self.port_groups = port_groups
        self.external_addresses = external_addresses
        self.networks = networks
        self.nat = nat

    @classmethod
    def create_from_configs(cls, config_path: str):
        """
        Load configuration from files
        """
        nat = NAT(config_path)
        return cls(
            secondary_configs.get_global_configuration(config_path),
            secondary_configs.get_port_groups(config_path),
            secondary_configs.get_external_addresses(config_path),
            [
                Network(
                    network_folder.split(path.sep)[-2],
                    nat,
                    config_path,
                    **(file_paths.load_yaml_from_file(network_folder))
                )
                for network_folder in file_paths.get_folders_with_config(
                    [config_path, file_paths.NETWORK_FOLDER]
                )
            ],
            nat,
        )

    def is_valid(self) -> bool:
        """
        Are all fields in the configuration valid
        """
        settings_valid = self.global_settings.validate()
        addresses_valid = self.external_addresses.validate()
        ports_valid = all([port.validate() for port in self.port_groups])
        networks_valid = all([network.validate() for network in self.networks])
        nat_valid = self.nat.validate()

        return (
            settings_valid
            and addresses_valid
            and ports_valid
            and networks_valid
            and nat_valid
        )

    def is_consistent(self) -> bool:
        """
        Check configuration for consistency
        """
        addresses_consistent = self.external_addresses.is_consistent()
        globals_consistent = self.global_settings.is_consistent()
        port_groups_consistent = [group.is_consistent() for group in self.port_groups]
        networks_consistent = [network.is_consistent() for network in self.networks]
        nat_consistent = self.nat.is_consistent()

        networks_consistent = all(networks_consistent) and True

        network_count = len(self.networks)
        for network_index in range(network_count):
            network = self.networks[network_index]

            # Skip any DHCP networks, e.g. for WAN
            if network.cidr is None:
                continue

            ip_network = ipaddress.ip_network(network.cidr)
            for second_network_index in range(network_index + 1, network_count):
                second_network = self.networks[second_network_index]

                # Skip any DHCP networks, e.g. for WAN
                if second_network.cidr is None:
                    continue

                second_ip_network = ipaddress.ip_network(second_network.cidr)

                if ip_network.overlaps(second_ip_network):
                    network.add_validation_error(
                        "{0} overlaps with {1}".format(
                            str(network), str(second_network)
                        )
                    )
                    networks_consistent = False

        return (
            addresses_consistent
            and globals_consistent
            and all(port_groups_consistent)
            and networks_consistent
            and nat_consistent
        )

    def validate(self) -> bool:
        """
        Is the root node valid
        """
        # Ensure both are checked, for complete error messages
        valid = self.is_valid()
        consistent = self.is_consistent()
        return valid and consistent

    def validation_failures(self) -> List[str]:
        """
        Get all validation failures
        """
        failures = (
            self.global_settings.validation_errors()
            + self.external_addresses.validation_errors()
            + self.nat.validation_failures()
        )
        for port in self.port_groups:
            failures.extend(port.validation_errors())
        for network in self.networks:
            failures.extend(network.validation_failures())

        return failures

    def get_commands(self) -> Tuple[List[List[str]], List[str]]:
        """
        Returns the commands to generate this configuration

        First value is an ordered list of commands to run, segmented into
        distinct portions which need to be committed in-order
        Second value is a flat list of all commands, for comparison against
        the previous configuration, to check for needed deletions
        """
        # These 3 should just be a list of commands, since ordering won't matter
        external_addresses = self.external_addresses.commands()
        port_groups = []
        for group in self.port_groups:
            port_groups.extend(group.commands())
        global_settings = self.global_settings.commands()
        nat_commands = self.nat.commands()

        # Address groups are used in NAT and firewall rules, which are set prior to
        # the host being parsed. To avoid chicken and egg problems, just pull out
        # the address groups here instead, which is sub-optimal for efficiency but
        # necessary to avoid breaking at runtime
        address_groups = []
        for network in self.networks:
            for host in network.hosts:
                if hasattr(host, "address-groups"):
                    address_groups.extend(
                        [
                            "firewall group address-group {0} address {1}".format(
                                group, host.address
                            )
                            for group in getattr(host, "address-groups")
                        ]
                    )

        ordered_commands = [
            [*external_addresses, *port_groups, *address_groups],
            global_settings,
            nat_commands,
        ]

        all_commands = external_addresses + port_groups + global_settings + nat_commands

        # Group ordered network commands together, extending the list of ordered
        # commands by all of them after they've all been created
        network_ordered_commands = []
        for network in self.networks:
            net_ordered_commands, net_command_list = network.commands()
            all_commands.extend(net_command_list)

            for index, commands in enumerate(net_ordered_commands):
                while index >= len(network_ordered_commands):
                    network_ordered_commands.append([])

                network_ordered_commands[index].extend(commands)

        ordered_commands.extend(network_ordered_commands)

        return (ordered_commands, all_commands)
