"""
Test interface
"""

from ubiquiti_config_generator import file_paths
from ubiquiti_config_generator.nodes import Interface
from ubiquiti_config_generator.testing_utils import counter_wrapper


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

    monkeypatch.setattr(Interface, "_add_keyword_attributes", fake_set_attrs)
    monkeypatch.setattr(Interface, "_load_firewalls", fake_load_firewalls)
    interface = Interface("interface", "network")

    assert interface.name == "interface", "Name set"
    assert interface.network_name == "network", "Network name set"
    assert fake_set_attrs.counter == 1, "Set attrs called"
    assert fake_load_firewalls.counter == 1, "Load firewalls called"

    assert str(interface) == "Interface interface", "Interface name returned"


def test_load_firewalls(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(
        file_paths,
        "get_folders_with_config",
        lambda folder: ["firewall1/config.yaml", "firewall2/config.yaml"],
    )
    monkeypatch.setattr(file_paths, "load_yaml_from_file", lambda file_path: {})

    interface = Interface("interface", "network")
    assert "firewalls" in interface._validate_attributes, "Firewalls added"
    firewalls = getattr(interface, "firewalls")
    assert "firewall1" in [firewall.name for firewall in firewalls], "Firewall 1 found"
    assert "firewall2" in [firewall.name for firewall in firewalls], "Firewall 2 found"
