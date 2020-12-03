"""
Configuration path definitions
"""
from os import path
from typing import List

TOP_LEVEL_DIRECTORY = "router_config"
GLOBAL_CONFIG = "global.yaml"
EXTERNAL_ADDRESSES_CONFIG = "external_addresses.yaml"
PORT_GROUPS_FOLDER = "port-groups"

NETWORK_FOLDER = "networks"
CONFIG_FILE_NAME = "config.yaml"
INTERFACE_FOLDER = "interfaces"
HOSTS_FOLDER = "hosts"
FIREWALL_FOLDER = "firewall"


def get_path(config_paths: List[str]):
    """
    Returns a file path with the top level directory prefixed
    """
    return path.abspath(path.join(".", TOP_LEVEL_DIRECTORY, *config_paths))
