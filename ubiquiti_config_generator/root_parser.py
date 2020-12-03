"""
Contains the root configuration node
"""
import glob
from os import path
from typing import List
import yaml

from ubiquiti_config_generator import file_paths
from ubiquiti_config_generator.nodes import PortGroup


def _get_global_configuration() -> dict:
    """
    Gets the yaml global configuration content
    """
    with open(file_paths.get_path(file_paths.GLOBAL_CONFIG)) as file_handle:
        return yaml.load(file_handle, Loader=yaml.FullLoader)


def _get_port_groups() -> List[PortGroup]:
    """
    Gets the yaml port group definitions
    """
    port_groups = []

    port_group_paths = glob.glob(
        path.join(file_paths.get_path(file_paths.PORT_GROUPS_FOLDER), "*.yaml")
    )
    for port_group in port_group_paths:
        group_name = path.basename(port_group).replace(".yaml", "")
        with open(port_group) as file_handle:
            ports = yaml.load(file_handle, Loader=yaml.FullLoader)

            port_groups.append(PortGroup(group_name, ports))


def _get_external_addresses() -> List[str]:
    """
    Gets the yaml external address definitions
    """
    with open(file_paths.get_path(file_paths.EXTERNAL_ADDRESSES_CONFIG)) as file_handle:
        return yaml.load(file_handle, Loader=yaml.FullLoader)


class RootNode:
    """
    Represents the root config node, from which everything else is based off of
    """
