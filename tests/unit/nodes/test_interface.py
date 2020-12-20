"""
Test interface
"""

from ubiquiti_config_generator import file_paths
from ubiquiti_config_generator.nodes import Interface, Firewall
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
    monkeypatch.setattr(
        file_paths, "load_yaml_from_file", lambda file_path: {"direction": "in"}
    )

    interface = Interface("interface", "network")
    assert "firewalls" in interface._validate_attributes, "Firewalls added"
    firewalls = getattr(interface, "firewalls")
    assert "firewall1" in [firewall.name for firewall in firewalls], "Firewall 1 found"
    assert "firewall2" in [firewall.name for firewall in firewalls], "Firewall 2 found"


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
    interface = Interface(
        "interface", "network", firewalls=[Firewall("firewall", "out")],
    )
    monkeypatch.setattr(Validatable, "validate", fake_validate)
    monkeypatch.setattr(Firewall, "validate", fake_validate)

    assert interface.validate(), "Interface is valid"
    assert fake_validate.counter == 2, "Called for parent/interface and firewall"


def test_validation_failures(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(file_paths, "get_folders_with_config", lambda folder: [])

    interface = Interface(
        "interface",
        "network",
        firewalls=[Firewall("firewall", "out"), Firewall("firewall2", "in")],
    )
    assert interface.validation_failures() == [], "No failures added yet"

    monkeypatch.setattr(Firewall, "validation_errors", lambda self: ["123"])
    interface.add_validation_error("failure")
    interface.add_validation_error("failure2")

    assert interface.validation_failures() == [
        "failure",
        "failure2",
        "123",
        "123",
    ], "Failures returned from interface and firewall"


def test_is_consistent(monkeypatch):
    """
    .
    """

    interface = Interface(
        "interface",
        "network",
        firewalls=[
            Firewall("firewall1", "out"),
            Firewall("firewall2", "in"),
            Firewall("firewall3", "local"),
            Firewall("firewall4", "in"),
        ],
    )

    assert not interface.is_consistent(), "Interface not consistent"
    assert interface.validation_errors() == [
        "Firewall firewall2 shares direction with Firewall firewall4"
    ], "Error added"
