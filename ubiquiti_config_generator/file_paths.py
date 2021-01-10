"""
Configuration path definitions
"""
import glob
from os import path
from typing import List, Union
import yaml

CURRENT_CONFIG_DIRECTORY = "router_config"
GLOBAL_CONFIG = "global_settings.yaml"
EXTERNAL_ADDRESSES_CONFIG = "external_addresses.yaml"
PORT_GROUPS_FOLDER = "port-groups"

NAT_FOLDER = "nat"
NETWORK_FOLDER = "networks"
CONFIG_FILE_NAME = "config.yaml"
HOSTS_FOLDER = "hosts"
FIREWALL_FOLDER = "firewalls"
RULE_FOLDER = "rules"


def load_yaml_from_file(file_path: str) -> Union[list, dict]:
    """
    Loads yaml data from a given file
    """
    with open(file_path) as file_handle:
        return yaml.load(file_handle, Loader=yaml.FullLoader)


def get_config_files(config_folders: List[str]) -> List[str]:
    """
    Returns a list of yaml files in a given directory
    """
    files = glob.glob(path.join(*config_folders, "*.yaml"))
    files.sort()
    return files


def get_folders_with_config(folder_paths: List[str]) -> List[str]:
    """
    Looks for config.yaml in nested folders under the provided one
    """
    folders = glob.glob(path.join(*folder_paths, "*", "config.yaml"))
    folders.sort()
    return folders


def get_path(config_paths: List[str]):
    """
    Returns a file path with the top level directory prefixed
    """
    return path.abspath(path.join(*config_paths))
