"""
Interactions for pushes from GitHub, triggering the deploy process
"""
import time
from typing import Optional

from ubiquiti_config_generator.github import api
from ubiquiti_config_generator.github.deployment_metadata import DeployMetadata
from ubiquiti_config_generator.messages import db
from ubiquiti_config_generator.messages.deployment import Deployment
from ubiquiti_config_generator.messages.log import Log


def check_push_for_deployment(
    deploy_config: dict, form: dict, access_token: str
) -> Optional[bool]:
    """
    Creates a deployment given a push action
    Returns true/false for deployment created or not, None for not applicable
    """
    # Immediately set commit status to pending, regardless of deploying or not
    # since it is pending checks
    api.set_commit_status(
        form["repository"]["statuses_url"].replace("{/sha}", "/" + form["after"]),
        form["after"],
        access_token,
        "pending",
    )

    metadata = DeployMetadata(form["before"], form["after"], None, None, deploy_config)

    return create_deployment(
        access_token, form["ref"], form["repository"]["deployments_url"], metadata
    )


def create_deployment(
    access_token: str, ref: str, deployment_url: str, metadata: DeployMetadata
) -> Optional[bool]:
    """
    Creates a deployment on GitHub and for local logging
    """
    if not is_against_primary_branch(metadata.deployment_configuration, ref):
        return None

    previous_commit = (
        api.get_active_deployment_sha(deployment_url, access_token)
        or metadata.before_sha
    )

    response = api.send_github_request(
        deployment_url,
        "post",
        access_token,
        {
            "ref": metadata.after_sha,
            "payload": {
                # Include the previously-deployed SHA, so we ensure that the deployment
                # can create the appropriate commands, diffing the active deployment
                # against the requested one
                "previous_commit": previous_commit
            },
        },
    )

    db.create_deployment(
        Deployment(previous_commit, metadata.after_sha, "requested", time.time())
    )

    if response.status_code != 201:
        print("Failed to create deployment")
        print(response.json())
        db.update_deployment_status(
            Log(
                metadata.after_sha,
                "Failed to create deployment",
                revision2=metadata.before_sha,
                status="failure",
            )
        )
    else:
        db.add_deployment_log(
            Log(
                metadata.after_sha,
                "Deployment created",
                revision2=metadata.before_sha,
                status="success",
            )
        )

    return response.status_code == 201


def is_against_primary_branch(deploy_config: dict, ref: str) -> bool:
    """
    Checks if the push is against the primary branch
    """
    return (
        ref.endswith("/" + deploy_config["git"]["primary-branch"])
        or ref == deploy_config["git"]["primary-branch"]
    )
