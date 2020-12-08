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

    GlobalSettings(stuff="things")
    assert fake_set_attrs.counter == 1, "Attributes set"
