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
    "name": type_checker.is_name,
    "cidr": type_checker.is_cidr,
    "authoritative": type_checker.is_string_boolean,
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

    def __init__(self, name: str, config_path: str, cidr: str, **kwargs):
        super().__init__(NETWORK_TYPES, ["name", "cidr"])
        self.name = name
        self.cidr = cidr
        self.config_path = config_path
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
                self.config_path,
                self.name,
                **(file_paths.load_yaml_from_file(interface_path))
            )
            for interface_path in file_paths.get_folders_with_config(
                file_paths.get_path(
                    [
                        self.config_path,
                        file_paths.NETWORK_FOLDER,
                        self.name,
                        file_paths.INTERFACE_FOLDER,
                    ]
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
                self.config_path,
                **(file_paths.load_yaml_from_file(host_path))
            )
            for host_path in file_paths.get_folders_with_config(
                file_paths.get_path(
                    [self.config_path, self.name, file_paths.HOSTS_FOLDER]
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
        all_commands = []
        ordered_commands = [[]]

        def append_command(command: str):
            all_commands.append(command)
            ordered_commands[-1].append(command)

        base = "service dhcp-server shared-network-name " + self.name
        if hasattr(self, "authoritative"):
            # pylint: disable=no-member
            append_command(base + " authoritative " + self.authoritative)

        subnet_base = base + " subnet " + self.cidr
        for subnet_attribute in [
            "domain-name",
            "default-router",
            "lease",
            "start",
            "dns-server",
        ]:
            if hasattr(self, subnet_attribute):
                append_command(
                    subnet_base
                    + " {0} {1}".format(
                        subnet_attribute, getattr(self, subnet_attribute)
                    )
                )

        if hasattr(self, "stop"):
            # pylint: disable=no-member
            append_command(
                subnet_base + " start {0} stop {1}".format(self.start, self.stop)
            )

        for server in getattr(self, "dns-servers", []):
            append_command(subnet_base + " dns-server " + server)

        for interface in self.interfaces:
            interface_commands, interface_command_list = interface.commands()
            all_commands.extend(interface_command_list)
            for commands in interface_commands:
                ordered_commands.extend(commands)

        mapping_base = subnet_base + " static-mapping "
        # First, set up address groups and static mappings
        for host in self.hosts:
            commands = []
            for group in getattr(host, "address-groups", []):
                commands.append(
                    "firewall group address-group {0} address {1}".format(
                        group, host.address
                    )
                )

            commands.extend(
                [
                    mapping_base + "{0} ip-address {1}".format(host.name, host.address),
                    mapping_base + "{0} mac-address {1}".format(host.name, host.mac),
                ]
            )
            host_commands, host_command_list = host.commands()

            ordered_commands.extend([commands, *host_commands])
            all_commands.extend([*commands, *host_command_list])

        return (ordered_commands, all_commands)

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "Network " + self.name
