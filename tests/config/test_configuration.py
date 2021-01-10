"""
Test configurations - loading, getting differences, etc
"""

from ubiquiti_config_generator import root_parser, nodes


def test_load_sample_config():
    """
    .
    """
    node = root_parser.RootNode.create_from_configs("sample_router_config")
    valid = node.validate()
    if not valid:
        [print(failure) for failure in node.validation_failures()]

    assert node.is_valid(), "Created node is valid"
    assert node.is_consistent(), "Created node is consistent"
    ordered_commands, command_list = node.get_commands()

    # external_addresses + port_groups + global_settings + nat_commands

    external_address = nodes.ExternalAddresses(["12.34.56.78", "23.45.67.89"])
    port_groups = [
        nodes.PortGroup(
            "server-ports", [80, 443, 22], "Ports to forward to the server"
        ),
        nodes.PortGroup("web", [80, 443], "Website ports"),
    ]

    settings = {
        "firewall/all-ping": "enable",
        "firewall/broadcast-ping": "disable",
        "system/time-zone": "America/New_York",
        "system/host-name": "router",
    }
    global_settings = nodes.GlobalSettings(**settings)

    masquerade_properties = {
        "description": "Masquerade",
        "log": "disable",
        "outbound-interface": "eth0",
        "type": "masquerade",
        "protocol": "all",
    }

    nat = nodes.NAT(".", [nodes.NATRule(5000, ".", **masquerade_properties)])

    drop_firewall = {"default-action": "drop"}
    accept_firewall = {"default-action": "accept"}

    admin_network = {
        "cidr": "10.0.10.0/23",
        "authoritative": "enable",
        "default-router": "10.0.10.1",
        "dns-server": "10.0.10.1",
        "dns-servers": ["8.8.8.8", "8.8.4.4"],
        "domain-name": "admin.home",
        "start": "10.0.11.1",
        "stop": "10.0.11.254",
        "lease": 86400,
        "interface_name": "eth1",
        "interface_description": "Internal interface",
        "vif": 10,
        "duplex": "auto",
        "speed": "auto",
        "firewalls": [
            nodes.Firewall(
                "administrative-IN",
                "in",
                "administrative",
                ".",
                description="Administrative inbound firewall",
                **drop_firewall,
                rules=[
                    nodes.Rule(
                        1,
                        "administrative-IN",
                        ".",
                        description="Allow established/related",
                        action="accept",
                        log="disable",
                        protocol="all",
                        state={"related": "enable", "established": "enable"},
                    )
                ]
            ),
            nodes.Firewall(
                "administrative-LOCAL",
                "local",
                "administrative",
                ".",
                description="Administrative local firewall",
                **accept_firewall,
                rules=[]
            ),
            nodes.Firewall(
                "administrative-OUT",
                "out",
                "administrative",
                ".",
                description="Administrative outbound firewall",
                **drop_firewall,
                rules=[
                    nodes.Rule(
                        1,
                        "administrative-OUT",
                        ".",
                        description="Allow established/related",
                        action="accept",
                        log="disable",
                        protocol="all",
                        state={"related": "enable", "established": "enable"},
                    )
                ]
            ),
        ],
    }

    internal_network = {
        "cidr": "10.0.12.0/24",
        "authoritative": "enable",
        "default-router": "10.0.12.1",
        "dns-server": "10.0.12.1",
        "dns-servers": ["8.8.8.8", "8.8.4.4"],
        "domain-name": "internal.home",
        "start": "10.0.12.100",
        "stop": "10.0.12.254",
        "lease": 86400,
        "interface_name": "eth1",
        "vif": 20,
        "duplex": "auto",
        "speed": "auto",
        "firewalls": [
            nodes.Firewall(
                "internal-IN",
                "in",
                "internal",
                ".",
                description="Internal inbound firewall",
                **drop_firewall,
                rules=[
                    nodes.Rule(
                        1,
                        "internal-IN",
                        ".",
                        description="Allow established/related",
                        action="accept",
                        log="disable",
                        protocol="all",
                        state={"related": "enable", "established": "enable"},
                    )
                ]
            ),
            nodes.Firewall(
                "internal-OUT",
                "out",
                "internal",
                ".",
                description="Internal outbound firewall",
                **drop_firewall,
                rules=[
                    nodes.Rule(
                        1,
                        "internal-OUT",
                        ".",
                        description="Allow established/related",
                        action="accept",
                        log="disable",
                        protocol="all",
                        state={"related": "enable", "established": "enable"},
                    )
                ]
            ),
        ],
    }

    untrusted_network = {
        "cidr": "10.200.0.0/24",
        "authoritative": "disable",
        "default-router": "10.200.0.1",
        "dns-server": "10.200.0.1",
        "dns-servers": ["8.8.8.8", "8.8.4.4"],
        "domain-name": "untrusted.home",
        "start": "10.200.0.100",
        "stop": "10.200.0.254",
        "lease": 3600,
        "interface_name": "eth2",
        "interface_description": "Untrusted interface",
        "duplex": "auto",
        "speed": "auto",
    }

    networks = [
        nodes.Network("administrative", nat, ".", **admin_network),
        nodes.Network("internal", nat, ".", **internal_network),
        nodes.Network("untrusted", nat, ".", **untrusted_network),
    ]

    rack_switch = nodes.Host(
        "rack-switch",
        networks[0],
        ".",
        **{
            "address": "10.0.10.2",
            "mac": "ba:21:dc:43:fe:65",
            "address-groups": ["infrastructure"],
            "forward-ports": {8081: 80},
            "connections": [
                {
                    "allow": False,
                    "destination": {"address": "10.0.10.2"},
                    "source": {"address": "10.200.0.0/24"},
                }
            ],
        }
    )

    command_list_pointer = 0
    # Rather than checking all 100+ commands explicitly, leverage unit tests to ensure
    # the commands of each class match the full list, one at a time
    for command_segment in [
        external_address,
        *port_groups,
        global_settings,
        nat,
        *networks,
    ]:
        commands = command_segment.commands()
        # Only care about unordered commands for now
        if isinstance(commands, tuple):
            commands = commands[1]

        end_commands = command_list_pointer + len(commands)
        assert (
            command_list[command_list_pointer:end_commands] == commands
        ), "Command segment correct for " + str(command_segment)
        command_list_pointer = end_commands

    assert command_list_pointer == len(command_list), "All commands checked"
