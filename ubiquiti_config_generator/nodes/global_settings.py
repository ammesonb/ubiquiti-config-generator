"""
Configurable global options
"""
import shlex
from typing import List

from ubiquiti_config_generator.nodes.validatable import Validatable


GLOBAL_SETTINGS_TYPES = {}


class GlobalSettings(Validatable):
    """
    Global options
    """

    def __init__(self, **kwargs):
        # Default validation to True if one is not provided for the given setting
        # Too many possibilities to exhaustively list them, so assume use knows what
        # they are doing in setting these
        for attr in kwargs:
            GLOBAL_SETTINGS_TYPES[attr] = GLOBAL_SETTINGS_TYPES.get(
                attr, lambda *args, **kwargs: True
            )

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

    def commands(self) -> List[str]:
        """
        Generate commands to set global settings
        """
        command_list = []
        for setting in self._validate_attributes:
            setting_path = setting.replace("/", " ")
            # For lists, make sure we preserve all values
            # Otherwise, just set the value directly
            if isinstance(getattr(self, setting), list):
                command_list.extend(
                    [
                        f"{setting_path} " + shlex.quote(str(value))
                        for value in getattr(self, setting)
                    ]
                )
            else:
                setting_value = shlex.quote(str(getattr(self, setting)))
                command_list.append(f"{setting_path} {setting_value}")

        return command_list
