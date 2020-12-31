"""
Test network
"""

from ubiquiti_config_generator import file_paths
from ubiquiti_config_generator.nodes import Network, Firewall, Host
from ubiquiti_config_generator.nodes.validatable import Validatable
from ubiquiti_config_generator.testing_utils import counter_wrapper

# pylint: disable=protected-access


def test_initialization(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_set_attrs(self, attrs: dict):
        """
        .
        """

    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_load_firewalls(self):
        """
        .
        """
        self.firewalls = []

    monkeypatch.setattr(Network, "_add_keyword_attributes", fake_set_attrs)
    monkeypatch.setattr(Network, "_load_firewalls", fake_load_firewalls)
    network = Network("network", ".", "10.0.0.0/8", "eth0")

    assert network.name == "network", "Name set"
    assert network.cidr == "10.0.0.0/8", "CIDR set"
    assert fake_set_attrs.counter == 1, "Set attrs called"
    assert fake_load_firewalls.counter == 1, "Load firewalls called"

    assert str(network) == "Network network", "Network name returned"


def test_load_firewalls(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(
        file_paths,
        "get_folders_with_config",
        lambda folder: ["firewall1/config.yaml", "firewall2/config.yaml"],
    )
    monkeypatch.setattr(Network, "_load_hosts", lambda self: None)
    # Need the direction set here for firewalls
    monkeypatch.setattr(
        file_paths,
        "load_yaml_from_file",
        lambda file_path: {"direction": "local" if "1" in file_path else "out"},
    )

    network = Network("network", ".", "10.0.0.0/8", "eth0")
    assert "firewalls" in network._validate_attributes, "firewalls added"
    firewalls = getattr(network, "firewalls")
    assert "firewall1" in [firewall.name for firewall in firewalls], "firewall 1 found"
    assert "firewall2" in [firewall.name for firewall in firewalls], "firewall 2 found"
    assert (
        network.firewalls_by_direction.get("in", None).name == "network-IN"
    ), "In firewall auto-constructed"
    assert (
        network.firewalls_by_direction.get("local", None).name == "firewall1"
    ), "Local firewall correct"
    assert (
        network.firewalls_by_direction.get("out", None).name == "firewall2"
    ), "Out firewall correct"


def test_load_hosts(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(
        file_paths,
        "get_folders_with_config",
        lambda folder: ["host1/config.yaml", "host2/config.yaml"],
    )
    monkeypatch.setattr(
        file_paths, "load_yaml_from_file", lambda file_path: {"address": "192.168.0.1"}
    )

    network = Network("network", ".", "10.0.0.0/8", "eth0", firewalls=[])
    assert "hosts" in network._validate_attributes, "Hosts added"
    hosts = getattr(network, "hosts")
    assert "host1" in [host.name for host in hosts], "Host 1 found"
    assert "host2" in [host.name for host in hosts], "Host 2 found"


def test_validate(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_validate(self):
        """
        .
        """
        return True

    monkeypatch.setattr(file_paths, "get_folders_with_config", lambda folder: [])
    network = Network(
        "network",
        ".",
        "1.1.1.1/24",
        "eth0",
        firewalls=[Firewall("firewall", "local", "network", ".")],
        hosts=[Host("host", None, ".", "192.168.0.1")],
    )
    monkeypatch.setattr(Validatable, "validate", fake_validate)
    monkeypatch.setattr(Firewall, "validate", fake_validate)
    monkeypatch.setattr(Host, "validate", fake_validate)

    assert network.validate(), "Network is valid"
    assert (
        fake_validate.counter == 5
    ), "Called for parent/network, all three firewalls, and host"


def test_validation_failures(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(file_paths, "get_folders_with_config", lambda folder: [])

    network = Network(
        "network",
        ".",
        "1.1.1.1/24",
        "eth0",
        firewalls=[Firewall("firewall", "in", "network", ".")],
        hosts=[
            Host("host", None, ".", "192.168.0.1"),
            Host("host2", None, ".", "192.168.0.1"),
        ],
    )
    assert network.validation_failures() == [], "No failures added yet"

    monkeypatch.setattr(Firewall, "validation_failures", lambda self: ["abc", "def"])
    monkeypatch.setattr(Host, "validation_errors", lambda self: ["ghi"])
    network.add_validation_error("failure")

    assert network.validation_failures() == [
        "failure",
        # Firewall failures repeated three times, one for each
        "abc",
        "def",
        "abc",
        "def",
        "abc",
        "def",
        "ghi",
        "ghi",
    ], "Failures returned for network and all children"


def test_is_consistent(monkeypatch):
    """
    .
    """

    @counter_wrapper
    def fake_host_consistent(self):
        """
        .
        """
        return True

    monkeypatch.setattr(Host, "is_consistent", fake_host_consistent)

    host1 = Host("test", None, ".", address="11.0.0.1", mac="ab:cd:ef")
    host2 = Host("test2", None, ".", address="10.0.0.1", mac="ab:cd:ef")
    host3 = Host("test3", None, ".", address="11.0.0.1", mac="ab:cd:12")
    host4 = Host("test4", None, ".", address="10.0.0.2", mac="12:34:56")

    firewall1 = Firewall("firewall", "in", "network", ".")
    firewall2 = Firewall("firewall2", "out", "network", ".")

    network_properties = {
        "firewalls": [firewall1, firewall2],
        "hosts": [host1, host2, host3, host4],
        "default-router": "192.168.0.1",
        "start": "10.10.0.100",
        "stop": "10.10.0.255",
        "interface_name": "eth0",
    }
    network = Network("network", ".", "10.0.0.0/24", **network_properties)

    assert not network.is_consistent(), "Network is not consistent"
    assert network.validation_errors() == [
        "Default router not in Network network",
        "DHCP start address not in Network network",
        "DHCP stop address not in Network network",
        "Host test not in Network network",
        "Host test3 not in Network network",
        "Host test shares an address with: Host test3",
        "Host test shares its mac with: Host test2",
    ], "Validation errors set"

    assert fake_host_consistent.counter == 4, "Four hosts' consistency checked"

    network_properties = {
        "firewalls": [firewall1, firewall2],
        "hosts": [host2, host4],
        "default-router": "10.0.0.1",
        "start": "10.0.0.100",
        "stop": "10.0.0.255",
        "interface_name": "eth0",
    }
    network = Network("network", ".", "10.0.0.0/24", **network_properties)
    assert network.is_consistent(), "Network should be consistent"
    assert not network.validation_errors(), "No validation errors present"


def test_network_commands(monkeypatch):
    """
    .
    """
    network_properties = {
        "authoritative": "disable",
        "domain-name": "test.domain",
        "default-router": "192.168.0.1",
        "lease": 86400,
        "start": "192.168.0.100",
        "stop": "192.168.0.254",
        "dns-server": "192.168.0.1",
        "dns-servers": ["8.8.8.8", "8.8.4.4"],
        "hosts": [],
        "firewalls": [],
    }
    monkeypatch.setattr(Firewall, "commands", lambda self: ([[]], []))
    monkeypatch.setattr(Host, "commands", lambda self: ([[]], []))
    network = Network("network1", ".", "192.168.0.0/24", "eth0", **network_properties)
    ordered_commands, command_list = network.commands()

    base = "service dhcp-server shared-network-name network1 "
    subnet_base = base + "subnet 192.168.0.0/24 "

    expected_commands = [
        base + "authoritative disable",
        subnet_base + "domain-name test.domain",
        subnet_base + "default-router 192.168.0.1",
        subnet_base + "lease 86400",
        subnet_base + "start 192.168.0.100",
        subnet_base + "dns-server 192.168.0.1",
        subnet_base + "start 192.168.0.100 stop 192.168.0.254",
        subnet_base + "dns-server 8.8.8.8",
        subnet_base + "dns-server 8.8.4.4",
        # This is auto-generated so must be included
        "interfaces ethernet eth0 address 192.168.0.1/24",
    ]

    assert command_list == expected_commands + [
        "interfaces ethernet eth0 firewall in name network1-IN",
        "interfaces ethernet eth0 firewall out name network1-OUT",
        "interfaces ethernet eth0 firewall local name network1-LOCAL",
    ], "Command list correct"
    assert ordered_commands == [
        expected_commands,
        [],  # Firewall commands would go here - checked in later test
        [
            "interfaces ethernet eth0 firewall in name network1-IN",
            "interfaces ethernet eth0 firewall out name network1-OUT",
            "interfaces ethernet eth0 firewall local name network1-LOCAL",
        ],
    ], "Ordered commands correct"


def test_interface_commands(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(Firewall, "commands", lambda self: ([[]], []))
    monkeypatch.setattr(Host, "commands", lambda self: ([[]], []))

    network_properties = {
        "hosts": [],
        "firewalls": [],
        "interface_description": "desc",
        "duplex": "auto",
        "speed": "auto",
    }
    network = Network("network1", ".", "192.168.0.0/24", "eth0", **network_properties)
    ordered_commands, command_list = network.commands()

    interface_base = "interfaces ethernet eth0 "
    expected_commands = [
        interface_base + "duplex auto",
        interface_base + "speed auto",
        interface_base + "address dhcp",
        interface_base + "description 'desc'",
    ]
    assert command_list == [
        *expected_commands,
        interface_base + "firewall in name network1-IN",
        interface_base + "firewall out name network1-OUT",
        interface_base + "firewall local name network1-LOCAL",
    ], "Command list for non-VIF interface correct"
    assert ordered_commands == [
        expected_commands,
        [],  # Firewall commands empty
        [
            interface_base + "firewall in name network1-IN",
            interface_base + "firewall out name network1-OUT",
            interface_base + "firewall local name network1-LOCAL",
        ],
    ], "Ordered commands correct"

    network_properties = {
        "hosts": [],
        "firewalls": [],
        "interface_description": "VLAN 123",
        "default-router": "192.168.0.1",
        "duplex": "half",
        "speed": 100,
        "vif": "123",
    }
    network = Network("network1", ".", "192.168.0.0/24", "eth0", **network_properties)
    ordered_commands, command_list = network.commands()

    interface_base = "interfaces ethernet eth0 "
    expected_commands = [
        "service dhcp-server shared-network-name network1 subnet 192.168.0.0/24 "
        "default-router 192.168.0.1",
        interface_base + "duplex half",
        interface_base + "speed 100",
        interface_base + "description 'CARRIER'",
        interface_base + "vif 123 address 192.168.0.1/24",
        interface_base + "vif 123 description 'VLAN 123'",
    ]
    assert command_list == [
        *expected_commands,
        interface_base + "vif 123 firewall in name network1-IN",
        interface_base + "vif 123 firewall out name network1-OUT",
        interface_base + "vif 123 firewall local name network1-LOCAL",
    ], "Command list for VIF interface correct"
    assert ordered_commands == [
        expected_commands,
        [],  # Firewall commands empty
        [
            interface_base + "vif 123 firewall in name network1-IN",
            interface_base + "vif 123 firewall out name network1-OUT",
            interface_base + "vif 123 firewall local name network1-LOCAL",
        ],
    ]


def test_command_ordering(monkeypatch):
    """
    .
    """

    @counter_wrapper
    def get_firewall_commands(self):
        """
        .
        """
        if get_firewall_commands.counter == 1:
            return (
                [["firewall1-command"], ["firewall1-command2"]],
                ["firewall1-command", "firewall1-command2"],
            )
        elif get_firewall_commands.counter == 2:
            return (
                [["firewall2-command", "firewall2-command2"], ["firewall2-command3"]],
                ["firewall2-command", "firewall2-command2", "firewall2-command3"],
            )
        else:
            return ([["firewall3-command"]], ["firewall3-command"])

    @counter_wrapper
    def get_host_commands(self):
        """
        .
        """
        if get_host_commands.counter == 1:
            return (
                [
                    ["host1-command", "host1-command2"],
                    ["host1-command3"],
                    ["host1-command4"],
                ],
                ["host1-command", "host1-command2", "host1-command3", "host1-command4"],
            )
        else:
            return (
                [["host2-command", "host2-command2"], ["host2-command3"]],
                ["host2-command", "host2-command2", "host2-command3"],
            )

    monkeypatch.setattr(Firewall, "commands", get_firewall_commands)
    monkeypatch.setattr(Host, "commands", get_host_commands)

    host_properties = {
        "mac": "abc",
        "address": "123",
        "address-groups": ["desktop", "windows"],
    }
    network_properties = {
        "domain-name": "test.domain",
        "default-router": "192.168.0.1",
        "lease": 86400,
        "start": "192.168.0.100",
        "stop": "192.168.0.254",
        "interface_description": "the interface",
        "hosts": [
            Host("host1", None, ".", **host_properties),
            Host("host2", None, ".", mac="def", address="234"),
        ],
        "firewalls": [
            Firewall("firewall1", "in", "network", "."),
            Firewall("firewall2", "out", "network", "."),
        ],
    }
    network = Network("network1", ".", "192.168.0.0/24", "eth0", **network_properties)
    ordered_commands, command_list = network.commands()

    base = "service dhcp-server shared-network-name network1 "
    subnet_base = base + "subnet 192.168.0.0/24 "
    mapping_base = subnet_base + "static-mapping "
    interface_base = "interfaces ethernet eth0 "

    assert command_list == [
        subnet_base + "domain-name test.domain",
        subnet_base + "default-router 192.168.0.1",
        subnet_base + "lease 86400",
        subnet_base + "start 192.168.0.100",
        subnet_base + "start 192.168.0.100 stop 192.168.0.254",
        interface_base + "address 192.168.0.1/24",
        interface_base + "description 'the interface'",
        "firewall1-command",
        "firewall1-command2",
        interface_base + "firewall in name firewall1",
        "firewall2-command",
        "firewall2-command2",
        "firewall2-command3",
        interface_base + "firewall out name firewall2",
        "firewall3-command",
        interface_base + "firewall local name network1-LOCAL",
        mapping_base + "host1 ip-address 123",
        mapping_base + "host1 mac-address abc",
        "firewall group address-group desktop address 123",
        "firewall group address-group windows address 123",
        "host1-command",
        "host1-command2",
        "host1-command3",
        "host1-command4",
        mapping_base + "host2 ip-address 234",
        mapping_base + "host2 mac-address def",
        "host2-command",
        "host2-command2",
        "host2-command3",
    ], "Network commands correct"

    assert ordered_commands == [
        [
            subnet_base + "domain-name test.domain",
            subnet_base + "default-router 192.168.0.1",
            subnet_base + "lease 86400",
            subnet_base + "start 192.168.0.100",
            subnet_base + "start 192.168.0.100 stop 192.168.0.254",
            interface_base + "address 192.168.0.1/24",
            interface_base + "description 'the interface'",
        ],
        [
            "firewall1-command",
            "firewall2-command",
            "firewall2-command2",
            "firewall3-command",
        ],
        ["firewall1-command2", "firewall2-command3"],
        [
            "interfaces ethernet eth0 firewall in name firewall1",
            "interfaces ethernet eth0 firewall out name firewall2",
            "interfaces ethernet eth0 firewall local name network1-LOCAL",
        ],
        [
            mapping_base + "host1 ip-address 123",
            mapping_base + "host1 mac-address abc",
            "firewall group address-group desktop address 123",
            "firewall group address-group windows address 123",
            mapping_base + "host2 ip-address 234",
            mapping_base + "host2 mac-address def",
        ],
        ["host1-command", "host1-command2", "host2-command", "host2-command2"],
        ["host1-command3", "host2-command3"],
        ["host1-command4"],
    ], "Ordered network commands correct"
