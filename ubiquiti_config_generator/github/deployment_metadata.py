"""
Helper config for deployment
"""
from dataclasses import dataclass


@dataclass
class DeployMetadata:
    """
    Represents the metadata for a deployment
    SHA for the before/after, associated URLs, etc
    """

    before_sha: str
    after_sha: str
    external_app_url: str
    status_url: str
    deployment_configuration: dict
