"""
Logging stuff
"""
from datetime import datetime, timezone
from typing import Optional

from ubiquiti_config_generator.web import print_utils


# pylint: disable=too-few-public-methods
class Log:
    """
    Represents an event log
    """

    # pylint: disable=too-many-arguments
    def __init__(
        self,
        revision1: str,
        message: str,
        utc_unix_timestamp: Optional[float],
        revision2: Optional[str] = None,
        status: Optional[str] = "log",
    ):
        self.revision1 = revision1
        self.message = message
        self.utc_unix_timestamp = utc_unix_timestamp or float(
            datetime.utcnow().astimezone(timezone.utc).strftime("%s.%f")
        )
        self.revision2 = revision2
        self.status = status

    def html(self) -> str:
        """
        Format this entry as HTML text
        """
        return [
            f"{print_utils.format_timestamp(self.utc_unix_timestamp)}",
            (
                f'[<span style="color: {print_utils.get_color_for_status(self.status)}">'
                + self.status
                + "</span>]"
            ),
            self.message,
        ]
