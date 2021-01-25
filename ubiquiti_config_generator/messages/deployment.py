"""
A summary about a deployment, with logs
"""
from typing import Optional, List

from ubiquiti_config_generator.messages.log import Log


# pylint: disable=too-few-public-methods
class Deployment:
    """
    Represents a deployment of a commit difference
    """

    # pylint: disable=too-many-arguments
    def __init__(
        self,
        from_revision: str,
        to_revision: str,
        status: str,
        started_at: float,
        ended_at: Optional[float] = None,
        logs: List[Log] = None,
    ):
        self.from_revision = from_revision
        self.to_revision = to_revision
        self.status = status
        self.started_at = started_at
        self.ended_at = ended_at
        self.logs = logs or []

    def __eq__(self, other) -> bool:
        """
        Is this equal to something else
        """
        return (
            isinstance(self, type(other))
            and self.from_revision == other.from_revision
            and self.to_revision == other.to_revision
            and self.status == other.status
            and self.started_at == other.started_at
            and self.ended_at == other.ended_at
            and self.logs == other.logs
        )
