"""
Contains the root configuration node
"""
from os import path
from typing import List

from ubiquiti_config_generator import file_paths
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

# TODO: this is a list of keys and possible values, in the form of a validator function


def _get_global_configuration() -> GlobalSettings:
    """
    Gets the yaml global configuration content
    """
    return GlobalSettings(
        **(
            file_paths.load_yaml_from_file(
                file_paths.get_path(file_paths.GLOBAL_CONFIG)
            )
        )
    )


def _get_port_groups() -> List[PortGroup]:
    """
    Gets the yaml port group definitions
    """
    port_groups = []
    for port_group in file_paths.get_config_files(file_paths.PORT_GROUPS_FOLDER):
        group_name = path.basename(port_group).replace(".yaml", "")
        ports = file_paths.load_yaml_from_file(port_group)
        port_groups.append(PortGroup(group_name, ports))

    return port_groups


def _get_external_addresses() -> ExternalAddresses:
    """
    Gets the yaml external address definitions
    """
    return ExternalAddresses(
        file_paths.load_yaml_from_file(
            file_paths.get_path(file_paths.EXTERNAL_ADDRESSES_CONFIG)
        ).get("addresses", [])
    )


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
            _get_global_configuration(),
            _get_port_groups(),
            _get_external_addresses(),
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
