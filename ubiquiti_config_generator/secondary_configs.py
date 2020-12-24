"""
Get secondary configurations, e.g. port groups, external addresses, etc
"""
from os import path
from typing import List

from ubiquiti_config_generator import file_paths
from ubiquiti_config_generator.nodes import (
    GlobalSettings,
    PortGroup,
    ExternalAddresses,
)


def get_global_configuration(config_path: str) -> GlobalSettings:
    """
    Gets the yaml global configuration content
    """
    return GlobalSettings(
        **(
            file_paths.load_yaml_from_file(
                file_paths.get_path([config_path, file_paths.GLOBAL_CONFIG])
            )
        )
    )


def get_port_groups(config_path: str) -> List[PortGroup]:
    """
    Gets the yaml port group definitions
    """
    port_groups = []
    for port_group in file_paths.get_config_files(
        [config_path, file_paths.PORT_GROUPS_FOLDER]
    ):
        group_name = path.basename(port_group).replace(".yaml", "")
        ports = file_paths.load_yaml_from_file(port_group)
        port_groups.append(PortGroup(group_name, ports))

    return port_groups


def get_external_addresses(config_path: str) -> ExternalAddresses:
    """
    Gets the yaml external address definitions
    """
    return ExternalAddresses(
        file_paths.load_yaml_from_file(
            file_paths.get_path([config_path, file_paths.EXTERNAL_ADDRESSES_CONFIG])
        ).get("addresses", [])
    )
