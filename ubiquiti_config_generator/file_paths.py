"""
Configuration path definitions
"""
import glob
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


def get_folders_with_config(folder_paths: List[str]) -> List[str]:
    """
    Looks for config.yaml in nested folders under the provided one
    """
    return glob.glob(path.join(get_path(folder_paths), "*", "*.yaml"))


def get_config_files(config_folders: List[str]) -> List[str]:
    """
    Returns a list of yaml files in a given directory
    """
    return glob.glob(path.join(get_path(config_folders), "*.yaml"))


def get_path(config_paths: List[str]):
    """
    Returns a file path with the top level directory prefixed
    """
    return path.abspath(path.join(".", TOP_LEVEL_DIRECTORY, *config_paths))
