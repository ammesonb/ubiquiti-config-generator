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
        "interface-name": "eth1",
        "interface-description": "Internal interface",
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
        "interface-name": "eth1",
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
        "interface-name": "eth2",
        "interface-description": "Untrusted interface",
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
            "forward-ports": [{8081: 80}],
            "connections": [
                {
                    "allow": False,
                    "description": "Disallow all connections to switch from "
                    "untrusted network",
                    "destination": {"address": "10.0.10.2"},
                    "source": {"address": "10.200.0.0/24"},
                }
            ],
        }
    )
    networks[0].hosts.append(rack_switch)
    server = nodes.Host(
        "server",
        networks[0],
        ".",
        **{
            "address": "10.0.10.10",
            "mac": "ab:12:cd:34:ef:56",
            "address-groups": ["infrastructure", "unix"],
            "forward-ports": ["server-ports", {8080: 80}],
            "hairpin-ports": [
                {
                    "description": "Redirect web to server",
                    "interface": "eth1.20",
                    "connection": {"destination": {"port": "web"}},
                },
                {
                    "description": "Redirect SSH to server",
                    "interface": "eth1.20",
                    "connection": {"destination": {"port": 22}},
                },
            ],
            "connections": [
                {
                    "allow": True,
                    "description": "Allow access to web ports",
                    "destination": {"address": "10.0.10.10", "port": "web"},
                },
                {
                    "allow": True,
                    "description": "Allow access to SSH from admin network",
                    "destination": {"address": "10.0.10.10", "port": "22"},
                    "source": {"address": "10.0.10.0/23"},
                },
                {
                    "allow": True,
                    "description": "Allow access to SSH from internal network",
                    "destination": {"address": "10.0.10.10", "port": "22"},
                    "source": {"address": "10.0.12.0/24"},
                },
                {
                    "allow": False,
                    "description": "Block all other access",
                    "log": True,
                    "destination": {"address": "10.0.10.10"},
                },
            ],
        }
    )
    networks[0].hosts.append(server)

    desktop = nodes.Host(
        "desktop",
        networks[1],
        ".",
        **{
            "address": "10.0.12.100",
            "mac": "ab:98:cd:65:ef:54",
            "address-groups": ["user-machines", "windows"],
            "connections": [
                {
                    "allow": True,
                    "description": "Allow connections to user devices from web "
                    "IOT ports",
                    "source": {"address": "10.200.0.0/24", "port": "web"},
                    "destination": {"address": "user-machines"},
                },
                {
                    "allow": False,
                    "description": "Block all other attempts to access user machines "
                    "from IOT",
                    "log": True,
                    "source": {"address": "10.200.0.0/24"},
                    "destination": {"address": "user-machines"},
                },
            ],
        }
    )
    networks[1].hosts.append(desktop)

    laptop = nodes.Host(
        "laptop",
        networks[1],
        ".",
        **{
            "address": "10.0.12.101",
            "mac": "fe:98:dc:76:ba:54",
            "address-groups": ["user-machines", "unix"],
        }
    )
    networks[1].hosts.append(laptop)

    echo = nodes.Host(
        "amazon-echo",
        networks[2],
        ".",
        **{
            "address": "10.200.0.10",
            "mac": "ab:12:bc:23:cd:34",
            "address-groups": ["IOT"],
            "connections": [
                {
                    "description": "Allow access to IOT from admin addresses",
                    "allow": True,
                    "log": False,
                    "source": {"address": "10.10.0.0/23"},
                    "destination": {"address": "IOT"},
                },
                {
                    "description": "Allow access to IOT from internal addresses",
                    "allow": True,
                    "log": False,
                    "source": {"address": "10.12.0.0/24"},
                    "destination": {"address": "IOT"},
                },
                {
                    "description": "Block access to IOT from others, unless "
                    "established already",
                    "allow": False,
                    "log": True,
                    "destination": {"address": "IOT"},
                },
            ],
        }
    )
    networks[2].hosts.append(echo)

    printer = nodes.Host(
        "printer",
        networks[2],
        ".",
        **{
            "address": "10.200.0.20",
            "mac": "fe:98:ed:87:dc:76",
            "address-groups": ["IOT"],
        }
    )
    networks[2].hosts.append(printer)

    teapot = nodes.Host(
        "teapot",
        networks[2],
        ".",
        **{
            "address": "10.200.0.30",
            "mac": "ad:14:be:25:cf:36",
            "address-groups": ["IOT"],
        }
    )
    networks[2].hosts.append(teapot)

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
