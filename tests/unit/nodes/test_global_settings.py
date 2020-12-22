"""
Test global settings
"""

from ubiquiti_config_generator.nodes import GlobalSettings
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_kw_set(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument
    @counter_wrapper
    def fake_set_attrs(self, kwargs: dict):
        """
        .
        """

    monkeypatch.setattr(GlobalSettings, "_add_keyword_attributes", fake_set_attrs)

    settings = GlobalSettings(stuff="things")
    assert fake_set_attrs.counter == 1, "Attributes set"
    assert str(settings) == "Global settings", "Name returned"


def test_is_consistent(monkeypatch):
    """
    .
    """
    # When there are more settings to test, update this
    settings = GlobalSettings()
    assert settings.is_consistent(), "Global settings should be consistent"


def test_commands():
    """
    .
    """
    attrs = {
        "firewall/all-ping": "enable",
        "system/ntp/server": "0.ubnt.pool.ntp.org",
        "system/host-name": "my router",
    }
    settings = GlobalSettings(**attrs)
    assert settings.commands() == [
        "firewall all-ping enable",
        "system ntp server 0.ubnt.pool.ntp.org",
        "system host-name 'my router'",
    ], "Global settings commands correct"
