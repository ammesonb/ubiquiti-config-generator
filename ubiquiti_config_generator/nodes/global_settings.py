"""
Configurable global options
"""

from ubiquiti_config_generator.nodes.validatable import Validatable


GLOBAL_SETTINGS_TYPES = {}


# Allow too few public methods, for now
# pylint: disable=too-few-public-methods


class GlobalSettings(Validatable):
    """
    Global options
    """

    def __init__(self, **kwargs):
        super().__init__(GLOBAL_SETTINGS_TYPES)
        self._add_keyword_attributes(kwargs)

    def __str__(self) -> str:
        """
        String version of this class
        """
        return "Global settings"

    def is_consistent(self) -> bool:
        """
        Are global settings internally consistent
        """
        # Nothing to actually validate here, yet
        return True
