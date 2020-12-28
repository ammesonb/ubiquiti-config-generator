"""
Test firewall
"""

from ubiquiti_config_generator.nodes import Firewall
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_firewall_calls_methods(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_set_attrs(self, attrs: dict):
        """
        .
        """

    monkeypatch.setattr(Firewall, "_add_keyword_attributes", fake_set_attrs)
    firewall = Firewall("firewall", "in", "network", ".")

    assert firewall.name == "firewall", "Name set"
    assert fake_set_attrs.counter == 1, "Set attrs called"

    assert str(firewall) == "Firewall firewall", "Name returned"
