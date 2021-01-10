"""
Contains the network node
"""
from os import path
import shlex
from typing import List, Tuple

from ubiquiti_config_generator import (
    file_paths,
    type_checker,
    utility,
)
from ubiquiti_config_generator.nodes import Firewall, Host, NAT
from ubiquiti_config_generator.nodes.validatable import Validatable

NETWORK_TYPES = {
    "name": type_checker.is_name,
    "cidr": type_checker.is_cidr,
    "authoritative": type_checker.is_string_boolean,
    # If not set, use DHCP for the interface
    "default-router": type_checker.is_ip_address,
    "dns-server": type_checker.is_ip_address,
    "dns-servers": lambda servers: all(
        [type_checker.is_ip_address(addr) for addr in servers]
    ),
    "domain-name": type_checker.is_string,
    "lease": type_checker.is_number,
    "start": type_checker.is_ip_address,
    "stop": type_checker.is_ip_address,
    # Interface properties
    "interface_name": type_checker.is_name,
    "interface_description": type_checker.is_description,
    "duplex": type_checker.is_duplex,
    "speed": type_checker.is_speed,
    "vif": type_checker.is_number,
    "firewalls": lambda firewalls: all([firewall.validate() for firewall in firewalls]),
    "hosts": lambda hosts: all([host.validate() for host in hosts]),
}


# pylint: disable=too-many-instance-attributes
class Network(Validatable):
    """
    A network to be created
    """

    # pylint: disable=too-many-arguments
    def __init__(
        self,
        name: str,
        nat: NAT,
        config_path: str,
        cidr: str,
        interface_name: str,
        **kwargs
    ):
        super().__init__(NETWORK_TYPES, ["name", "cidr", "firewalls", "hosts"])
        self.name = name
        self.nat = nat
        self.cidr = cidr
        self.config_path = config_path
        self.interface_name = interface_name
        self._add_keyword_attributes(kwargs)

        self.firewalls_by_direction = {}
        if "firewalls" not in kwargs:
            self._load_firewalls()

        for firewall in self.firewalls:
            if hasattr(firewall, "direction"):
                self.firewalls_by_direction[firewall.direction] = firewall

        # Add placeholder for any missing firewalls
        firewall_properties = {"default-action": "accept"}
        for firewall_direction in ["in", "out", "local"]:
            if firewall_direction not in self.firewalls_by_direction:
                new_firewall = Firewall(
                    self.name + "-" + firewall_direction.upper(),
                    firewall_direction,
                    self.name,
                    self.config_path,
                    **firewall_properties
                )
                self.firewalls.append(new_firewall)
                self.firewalls_by_direction[firewall_direction] = new_firewall

        if "hosts" not in kwargs:
            self._load_hosts()

    def _load_firewalls(self) -> None:
        """
        Load firewalls for this network
        """
        self.firewalls = [
            Firewall(
                firewall_path.split(path.sep)[-2],
                network_name=self.name,
                config_path=self.config_path,
                **(file_paths.load_yaml_from_file(firewall_path))
            )
            for firewall_path in file_paths.get_folders_with_config(
                [
                    self.config_path,
                    file_paths.NETWORK_FOLDER,
                    self.name,
                    file_paths.FIREWALL_FOLDER,
                ]
            )
        ]

    def _load_hosts(self) -> None:
        """
        Load hosts for this network
        """
        self.hosts = [
            Host(
                host_path.split(path.sep)[-1][:-5],
                self,
                self.config_path,
                **(file_paths.load_yaml_from_file(host_path))
            )
            for host_path in file_paths.get_config_files(
                [
                    self.config_path,
                    file_paths.NETWORK_FOLDER,
                    self.name,
                    file_paths.HOSTS_FOLDER,
                ]
            )
        ]

    def validation_failures(self) -> List[str]:
        """
        Get all validation failures
        """
        failures = self.validation_errors()
        for firewall in self.firewalls:
            failures.extend(firewall.validation_failures())
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

        hosts_consistent = [host.is_consistent() for host in self.hosts]
        return consistent and all(hosts_consistent)

    # pylint: disable=too-many-locals,too-many-branches,too-many-statements
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

        # First set up basic properties of the subnet
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

        # Set DHCP stop range
        if hasattr(self, "stop"):
            # pylint: disable=no-member
            append_command(
                subnet_base + " start {0} stop {1}".format(self.start, self.stop)
            )

        # Set DNS servers
        for server in getattr(self, "dns-servers", []):
            append_command(subnet_base + " dns-server " + server)

        # Then add interface attributes
        interface_base = "interfaces ethernet {0}".format(self.interface_name)
        for interface_attribute in ["duplex", "speed"]:
            if hasattr(self, interface_attribute):
                append_command(
                    interface_base
                    + " {0} {1}".format(
                        interface_attribute, getattr(self, interface_attribute)
                    )
                )

        # If there is a VIF on the interface, mark as carrier
        if hasattr(self, "vif"):
            append_command(interface_base + " description 'CARRIER'")

        # Address/description should be set on the VIF if there is one
        # pylint: disable=no-member
        address_base = interface_base + (
            " vif {0} ".format(self.vif) if hasattr(self, "vif") else " "
        )
        if hasattr(self, "default-router"):
            append_command(
                address_base
                + "address "
                + getattr(self, "default-router")
                + "/"
                + str(self.cidr.split("/")[1])
            )
        else:
            append_command(address_base + "address dhcp")

        if hasattr(self, "interface_description"):
            # pylint: disable=no-member
            description = shlex.quote(self.interface_description)
            if description[0] not in ['"', "'"]:
                description = "'{0}'".format(description)

            append_command(address_base + "description " + description)

        # Get firewall commands and add them
        ordered_firewall_commands = []
        first_firewall = True
        for firewall in self.firewalls:
            firewall_commands, firewall_command_list = firewall.commands()
            all_commands.extend(firewall_command_list)

            command_index = 0
            for commands in firewall_commands:
                while len(ordered_firewall_commands) <= command_index:
                    ordered_firewall_commands.append([])
                ordered_firewall_commands[command_index].extend(commands)
                command_index += 1

            interface_firewall = address_base + "firewall {0} name {1}".format(
                firewall.direction, firewall.name
            )
            if first_firewall:
                ordered_firewall_commands.append([])
                first_firewall = False

            ordered_firewall_commands[-1].append(interface_firewall)
            all_commands.append(interface_firewall)

        ordered_commands.extend(ordered_firewall_commands)

        # Add address groups and static mappings for hosts
        # Plus the commands for the host itself
        ordered_host_commands = []
        mapping_base = subnet_base + " static-mapping "
        for host in self.hosts:
            host_commands = []
            host_command_list = []

            static_commands = [
                mapping_base + "{0} ip-address {1}".format(host.name, host.address),
                mapping_base + "{0} mac-address {1}".format(host.name, host.mac),
            ]

            for group in getattr(host, "address-groups", []):
                static_commands.append(
                    "firewall group address-group {0} address {1}".format(
                        group, host.address
                    )
                )

            # Add the static mapping commands to the beginning of the ordered list
            host_commands.insert(0, [])
            host_commands[0].extend(static_commands)
            # Set the flat list to have the static commands at the front of it
            host_command_list = [*static_commands, *host_command_list]

            command_index = 0
            for commands in host_commands:
                while len(ordered_host_commands) <= command_index:
                    ordered_host_commands.append([])
                ordered_host_commands[command_index].extend(commands)
                command_index += 1

            all_commands.extend(host_command_list)

        ordered_commands.extend(ordered_host_commands)

        return (ordered_commands, all_commands)

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "Network " + self.name
