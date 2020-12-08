"""
Test port group
"""

from ubiquiti_config_generator.nodes import PortGroup


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
