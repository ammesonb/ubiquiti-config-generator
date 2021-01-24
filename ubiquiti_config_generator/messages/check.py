"""
A summary about a check, with logs
"""
from dataclasses import dataclass
from typing import Optional, List

from ubiquiti_config_generator.messages.log import Log


@dataclass
class Check:
    """
    Represents a check of a commit
    """

    revision: str
    status: str
    started_at: float
    ended_at: Optional[float]
    logs: List[Log]
