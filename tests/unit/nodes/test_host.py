"""
Test host
"""

from ubiquiti_config_generator.nodes import Host
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_host_calls_methods(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_set_attrs(self, attrs: dict):
        """
        .
        """

    monkeypatch.setattr(Host, "_add_keyword_attributes", fake_set_attrs)
    host = Host("host")

    assert host.name == "host", "Name set"
    assert fake_set_attrs.counter == 1, "Set attrs called"

    assert str(host) == "Host host", "Name returned"
