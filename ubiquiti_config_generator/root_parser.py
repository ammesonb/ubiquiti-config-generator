"""
Contains the root configuration node
"""
import ipaddress
from os import path
from typing import List

from ubiquiti_config_generator import file_paths, secondary_configs
from ubiquiti_config_generator.nodes import (
    GlobalSettings,
    PortGroup,
    ExternalAddresses,
    Network,
)

# Allow TODO comments while WIP
# pylint: disable=fixme
# TODO: For commenting on issue with commands to run
# https://github.com/ActionsDesk/add-comment-action
# TODO: to get cache of previous configuration, make file changes like so:
# Can run bash directly - https://github.com/jpadfield/simple-site/blob/master/.github/workflows/build.yml
# Then use this to commit and push back:
# https://github.com/stefanzweifel/git-auto-commit-action


class RootNode:
    """
    Represents the root config node, from which everything else is based off of
    """

    def __init__(
        self,
        global_settings: GlobalSettings,
        port_groups: List[PortGroup],
        external_addresses: ExternalAddresses,
        networks: List[Network],
    ):
        self.global_settings = global_settings
        self.port_groups = port_groups
        self.external_addresses = external_addresses
        self.networks = networks

    @classmethod
    def create_from_configs(cls):
        """
        Load configuration from files
        """
        return cls(
            secondary_configs.get_global_configuration(),
            secondary_configs.get_port_groups(),
            secondary_configs.get_external_addresses(),
            [
                Network(
                    name=network_folder.split(path.sep)[-2],
                    **(file_paths.load_yaml_from_file(network_folder))
                )
                for network_folder in file_paths.get_folders_with_config(
                    file_paths.NETWORK_FOLDER
                )
            ],
        )

    def is_valid(self) -> bool:
        """
        Are all fields in the configuration valid
        """
        return (
            self.global_settings.validate()
            and self.external_addresses.validate()
            and all([port.validate() for port in self.port_groups])
            and all([network.validate() for network in self.networks])
        )

    def is_consistent(self) -> bool:
        """
        Check configuration for consistency
        """
        addresses_consistent = self.external_addresses.is_consistent()
        globals_consistent = self.global_settings.is_consistent()
        port_groups_consistent = [group.is_consistent() for group in self.port_groups]
        networks_consistent = [network.is_consistent() for network in self.networks]

        networks_consistent = all(networks_consistent) and True

        network_count = len(self.networks)
        for network_index in range(network_count):
            network = self.networks[network_index]
            ip_network = ipaddress.ip_network(network.cidr)
            for second_network_index in range(network_index + 1, network_count):
                second_network = self.networks[second_network_index]
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
        )

    def validate(self) -> bool:
        """
        Is the root node valid
        """
        return self.is_valid() and self.is_consistent()

    def validation_failures(self) -> List[str]:
        """
        Get all validation failures
        """
        failures = (
            self.global_settings.validation_errors()
            + self.external_addresses.validation_errors()
        )
        for port in self.port_groups:
            failures.extend(port.validation_errors())
        for network in self.networks:
            failures.extend(network.validation_failures())

        return failures

    def find_changes_from(self, previous_config: "RootNode"):
        """
        Using a previous configuration, identify what has changed and
        return a condensed summary suitable for applying the changes
        """
        # TODO: make an object capable of representing changes
        # TODO: how do we even _get_ the previous configuration in a single run of this?
        # TODO: then need to make commands to apply changes
