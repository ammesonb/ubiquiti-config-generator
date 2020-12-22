"""
Test port group
"""

from ubiquiti_config_generator.nodes import PortGroup
from ubiquiti_config_generator import utility


def test_port_group():
    """
    .
    """
    ports = PortGroup("group", [80])
    assert ports.name == "group", "Name set"
    assert ports.ports == [80], "Ports set"

    ports.add_port(443)
    assert ports.ports == [80, 443], "Port added"

    ports.add_ports([8080, 4430])
    assert ports.ports == [80, 443, 8080, 4430], "Ports added"

    assert str(ports) == "Port group group", "Port group name returned"


def test_is_consistent(monkeypatch):
    """
    .
    """
    monkeypatch.setattr(utility, "get_duplicates", lambda array: [])
    group = PortGroup("group", [])
    assert group.is_consistent(), "No duplicates is consistent"

    monkeypatch.setattr(utility, "get_duplicates", lambda array: [10, 20])

    duplicate_group = PortGroup("group2", [])
    assert not duplicate_group.is_consistent(), "Duplicate port group is not consistent"
    assert duplicate_group.validation_errors() == [
        str(duplicate_group) + " has duplicate ports: 10, 20"
    ], "Validation error added"


def test_command():
    """
    .
    """
    group = PortGroup("web-ports", [80, 443])
    group2 = PortGroup(
        "printer-ports", [161, 515, 631, 9100], "Ports for printer connections"
    )

    assert group.commands() == [
        "firewall group port-group web-ports port 80",
        "firewall group port-group web-ports port 443",
    ], "No description port group correct"

    assert group2.commands() == [
        "firewall group port-group printer-ports port 161",
        "firewall group port-group printer-ports port 515",
        "firewall group port-group printer-ports port 631",
        "firewall group port-group printer-ports port 9100",
        'firewall group port-group printer-ports description "Ports for printer connections"',
    ], "Description port group correct"
