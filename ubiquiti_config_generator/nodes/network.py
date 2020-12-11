"""
Contains the network node
"""
from os import path
from typing import List

from ubiquiti_config_generator import file_paths, type_checker
from ubiquiti_config_generator.nodes import Interface, Host
from ubiquiti_config_generator.nodes.validatable import Validatable

NETWORK_TYPES = {
    "name": type_checker.is_string,
    "subnet": type_checker.is_cidr,
    "default-router": type_checker.is_ip_address,
    "dns-server": type_checker.is_ip_address,
    "dns-servers": lambda servers: all(
        [type_checker.is_ip_address(addr) for addr in servers]
    ),
    "domain-name": type_checker.is_string,
    "lease": type_checker.is_number,
    "start": type_checker.is_ip_address,
    "stop": type_checker.is_ip_address,
    "interfaces": lambda interfaces: all(
        [interface.validate() for interface in interfaces]
    ),
}


class Network(Validatable):
    """
    A network to be created
    """

    def __init__(self, name: str, cidr: str, **kwargs):
        super().__init__(NETWORK_TYPES, ["name", "cidr"])
        self.name = name
        self.cidr = cidr
        self._add_keyword_attributes(kwargs)
        self._load_interfaces()
        self._load_hosts()

    def _load_interfaces(self) -> None:
        """
        Load interfaces for this network
        """
        self.interfaces = [
            Interface(
                interface_path.split(path.sep)[-2],
                self.name,
                **(file_paths.load_yaml_from_file(interface_path))
            )
            for interface_path in file_paths.get_folders_with_config(
                file_paths.get_path(
                    [file_paths.NETWORK_FOLDER, self.name, file_paths.INTERFACE_FOLDER]
                )
            )
        ]
        self._add_validate_attribute("interfaces")

    def _load_hosts(self) -> None:
        """
        Load hosts for this network
        """
        self.hosts = [
            Host(
                host_path.split(path.sep)[-2],
                **(file_paths.load_yaml_from_file(host_path))
            )
            for host_path in file_paths.get_folders_with_config(
                file_paths.get_path(
                    [file_paths.NETWORK_FOLDER, self.name, file_paths.HOSTS_FOLDER]
                )
            )
        ]
        self._add_validate_attribute("hosts")

    def validation_failures(self) -> List[str]:
        """
        Get all validation failures
        """
        failures = self.validation_errors()
        for interface in self.interfaces:
            failures.extend(interface.validation_errors())
        for host in self.hosts:
            failures.extend(host.validation_errors())
        return failures

    def is_consistent(self) -> bool:
        """
        Check configuration for consistency
        """

    def validate(self) -> bool:
        """
        Is the root node valid
        """
        return (
            super().validate()
            and all([interface.validate() for interface in self.interfaces])
            and all([host.validate() for host in self.hosts])
        )

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "Network " + self.name
