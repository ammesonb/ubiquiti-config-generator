"""
A firewall rule
"""
import shlex
from typing import List

from ubiquiti_config_generator import type_checker
from ubiquiti_config_generator.nodes.validatable import Validatable


RULE_TYPES = {
    "number": type_checker.is_number,
    "description": type_checker.is_description,
    "log": type_checker.is_string_boolean,
    "source": type_checker.is_address_and_or_port,
    "destination": type_checker.is_address_and_or_port,
    "protocol": type_checker.is_protocol,
    "inbound_interface": type_checker.is_string,
    "outbound_interface": type_checker.is_string,
    "type": type_checker.is_nat_type,
}


class NATRule(Validatable):
    """
    Represents a firewall rule
    """

    def __init__(self, number: int, firewall_name: str, **kwargs):
        super().__init__(RULE_TYPES, ["number"])
        self.number = number
        self.firewall_name = firewall_name
        self._add_keyword_attributes(kwargs)

    # pylint: disable=too-many-branches
    def commands(self) -> List[str]:
        """
        Get the command for this rule
        """
        command_base = "service nat rule {1} ".format(self.firewall_name, self.number)

        commands = [command_base + "action " + getattr(self, "action", "accept")]

        if hasattr(self, "description"):
            # pylint: disable=no-member
            description = shlex.quote(self.description)
            if description[0] not in ["'", '"']:
                description = '"{0}"'.format(description)

            commands.append(command_base + "description " + description)

        for part in ["log", "protocol"]:
            if hasattr(self, part):
                commands.append(command_base + part + " " + getattr(self, part))

        if hasattr(self, "state"):
            # pylint: disable=no-member
            for state, enabled in self.state.items():
                commands.append(command_base + "state {0} {1}".format(state, enabled))

        connections = ["source", "destination"]
        for connection in connections:
            if hasattr(self, connection):
                data = getattr(self, connection)

                if "address" in data:
                    if type_checker.is_ip_address(
                        data["address"]
                    ) or type_checker.is_cidr(data["address"]):
                        commands.append(
                            command_base + connection + " address " + data["address"]
                        )
                    else:
                        commands.append(
                            command_base
                            + connection
                            + " group address-group "
                            + data["address"]
                        )

                if "port" in data:
                    if type_checker.is_number(data["port"]):
                        commands.append(
                            command_base + connection + " port " + str(data["port"])
                        )
                    else:
                        commands.append(
                            command_base
                            + connection
                            + " group port-group "
                            + data["port"]
                        )

        return commands
