"""
Contains the network node
"""
from os import path
from typing import List, Tuple

from ubiquiti_config_generator import (
    file_paths,
    type_checker,
    utility,
)
from ubiquiti_config_generator.nodes import Interface, Host
from ubiquiti_config_generator.nodes.validatable import Validatable

NETWORK_TYPES = {
    "name": type_checker.is_string,
    "cidr": type_checker.is_cidr,
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
    "hosts": lambda hosts: all([host.validate() for host in hosts]),
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

        if "interfaces" not in kwargs:
            self._load_interfaces()
        if "hosts" not in kwargs:
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
            failures.extend(interface.validation_failures())
        for host in self.hosts:
            failures.extend(host.validation_errors())
        return failures

    def is_consistent(self) -> bool:
        """
        Check configuration for consistency
        """
        consistent = True

        if not utility.address_in_subnet(
            self.cidr, getattr(self, "default-router", None)
        ):
            self.add_validation_error("Default router not in " + str(self))
            consistent = False

        if not utility.address_in_subnet(self.cidr, getattr(self, "start", None)):
            self.add_validation_error("DHCP start address not in " + str(self))
            consistent = False

        if not utility.address_in_subnet(self.cidr, getattr(self, "stop", None)):
            self.add_validation_error("DHCP stop address not in " + str(self))
            consistent = False

        for host in self.hosts:
            if not utility.address_in_subnet(self.cidr, host.address):
                self.add_validation_error("{0} not in {1}".format(str(host), str(self)))
                consistent = False

        for first_host_index in range(len(self.hosts)):
            first_host = self.hosts[first_host_index]
            matched_addresses = [
                second_host
                for second_host in self.hosts[first_host_index + 1 :]
                if first_host.address == second_host.address
            ]
            matched_macs = [
                second_host
                for second_host in self.hosts[first_host_index + 1 :]
                if first_host.mac == second_host.mac
            ]
            if matched_addresses:
                self.add_validation_error(
                    "{0} shares an address with: {1}".format(
                        str(first_host),
                        ", ".join([str(host) for host in matched_addresses]),
                    )
                )
                consistent = False

            if matched_macs:
                self.add_validation_error(
                    "{0} shares its mac with: {1}".format(
                        str(first_host),
                        ", ".join([str(host) for host in matched_macs]),
                    )
                )
                consistent = False

        interfaces_consistent = [
            interface.is_consistent() for interface in self.interfaces
        ]
        hosts_consistent = [host.is_consistent() for host in self.hosts]
        return consistent and all(interfaces_consistent) and all(hosts_consistent)

    def validate(self) -> bool:
        """
        Is the root node valid
        """
        return (
            super().validate()
            and all([interface.validate() for interface in self.interfaces])
            and all([host.validate() for host in self.hosts])
        )

    def commands(self) -> Tuple[List[List[str]], List[str]]:
        """
        The commands to generate this network
        """
        return ([], [])

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "Network " + self.name
