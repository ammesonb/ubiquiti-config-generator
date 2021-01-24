"""
A summary about a deployment, with logs
"""
from dataclasses import dataclass
from typing import Optional, List

from ubiquiti_config_generator.messages.log import Log


@dataclass
class Deployment:
    """
    Represents a deployment of a commit difference
    """

    from_revision: str
    to_revision: str
    status: str
    started_at: float
    ended_at: Optional[float]
    logs: List[Log]
