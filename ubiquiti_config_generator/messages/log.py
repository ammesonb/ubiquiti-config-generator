"""
Logging stuff
"""
from dataclasses import dataclass
from typing import Optional


@dataclass
class Log:
    """
    Represents an event log
    """

    revision1: str
    message: str
    status: Optional[str] = "log"
    utc_unix_timestamp: Optional[float]
    revision2: Optional[str]
