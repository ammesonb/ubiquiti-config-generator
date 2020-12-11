"""
Test network
"""

from ubiquiti_config_generator import file_paths
from ubiquiti_config_generator.nodes import Network
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
    def fake_load_interfaces(self):
        """
        .
        """

    monkeypatch.setattr(Network, "_add_keyword_attributes", fake_set_attrs)
    monkeypatch.setattr(Network, "_load_interfaces", fake_load_interfaces)
    network = Network("network", "10.0.0.0/8")

    assert network.name == "network", "Name set"
    assert network.cidr == "10.0.0.0/8", "CIDR set"
    assert fake_set_attrs.counter == 1, "Set attrs called"
    assert fake_load_interfaces.counter == 1, "Load interfaces called"

    assert str(network) == "Network network", "Network name returned"


def test_load_interfaces(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(
        file_paths,
        "get_folders_with_config",
        lambda folder: ["interface1/config.yaml", "interface2/config.yaml"],
    )
    monkeypatch.setattr(file_paths, "load_yaml_from_file", lambda file_path: {})

    network = Network("network", "10.0.0.0/8")
    assert "interfaces" in network._validate_attributes, "Interfaces added"
    interfaces = getattr(network, "interfaces")
    assert "interface1" in [
        interface.name for interface in interfaces
    ], "Interface 1 found"
    assert "interface2" in [
        interface.name for interface in interfaces
    ], "Interface 2 found"


def test_load_hosts(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(
        file_paths,
        "get_folders_with_config",
        lambda folder: ["host1/config.yaml", "host2/config.yaml"],
    )
    monkeypatch.setattr(file_paths, "load_yaml_from_file", lambda file_path: {})

    network = Network("network", "10.0.0.0/8")
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


def test_validation_failures(monkeypatch):
    """
    .
    """
