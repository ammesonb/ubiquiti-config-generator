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

    monkeypatch.setattr(Network, "_add_keyword_attributes", fake_set_attrs)
    monkeypatch.setattr(Network, "_load_firewalls", fake_load_firewalls)
    network = Network("network", ".", "10.0.0.0/8")

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
        file_paths, "load_yaml_from_file", lambda file_path: {"direction": "local"}
    )

    network = Network("network", ".", "10.0.0.0/8")
    assert "firewalls" in network._validate_attributes, "firewalls added"
    firewalls = getattr(network, "firewalls")
    assert "firewall1" in [firewall.name for firewall in firewalls], "firewall 1 found"
    assert "firewall2" in [firewall.name for firewall in firewalls], "firewall 2 found"


def test_load_hosts(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(
        file_paths,
        "get_folders_with_config",
        lambda folder: ["host1/config.yaml", "host2/config.yaml"],
    )
    monkeypatch.setattr(Network, "_load_firewalls", lambda self: None)
    monkeypatch.setattr(file_paths, "load_yaml_from_file", lambda file_path: {})

    network = Network("network", ".", "10.0.0.0/8")
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
        firewalls=[Firewall("firewall", "local")],
        hosts=[Host("host", ".")],
    )
    monkeypatch.setattr(Validatable, "validate", fake_validate)
    monkeypatch.setattr(Firewall, "validate", fake_validate)
    monkeypatch.setattr(Host, "validate", fake_validate)

    assert network.validate(), "Network is valid"
    assert fake_validate.counter == 3, "Called for parent/network, firewall, and host"


def test_validation_failures(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(file_paths, "get_folders_with_config", lambda folder: [])

    network = Network(
        "network",
        ".",
        "1.1.1.1/24",
        firewalls=[Firewall("firewall", "in")],
        hosts=[Host("host", "."), Host("host2", ".")],
    )
    assert network.validation_failures() == [], "No failures added yet"

    monkeypatch.setattr(Firewall, "validation_errors", lambda self: ["abc", "def"])
    monkeypatch.setattr(Host, "validation_errors", lambda self: ["ghi"])
    network.add_validation_error("failure")

    assert network.validation_failures() == [
        "failure",
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

    host1 = Host("test", ".", address="11.0.0.1", mac="ab:cd:ef")
    host2 = Host("test2", ".", address="10.0.0.1", mac="ab:cd:ef")
    host3 = Host("test3", ".", address="11.0.0.1", mac="ab:cd:12")
    host4 = Host("test4", ".", address="10.0.0.2", mac="12:34:56")

    firewall1 = Firewall("firewall", "in")
    firewall2 = Firewall("firewall2", "out")

    network_properties = {
        "firewalls": [firewall1, firewall2],
        "hosts": [host1, host2, host3, host4],
        "default-router": "192.168.0.1",
        "start": "10.10.0.100",
        "stop": "10.10.0.255",
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
    }
    network = Network("network", ".", "10.0.0.0/24", **network_properties)
    assert network.is_consistent(), "Network should be consistent"
    assert not network.validation_errors(), "No validation errors present"
