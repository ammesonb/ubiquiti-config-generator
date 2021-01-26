"""
Interactions for pushes from GitHub, triggering the deploy process
"""
import time
from typing import Optional

from ubiquiti_config_generator.github import api
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

    if not is_against_primary_branch(deploy_config, form["ref"]):
        return None

    previous_commit = (
        api.get_active_deployment_sha(
            form["repository"]["deployments_url"], access_token
        )
        or form["before"]
    )

    db.create_deployment(
        Deployment(previous_commit, form["after"], "requested", time.time())
    )

    response = api.send_github_request(
        form["repository"]["deployments_url"],
        "post",
        access_token,
        {
            "ref": form["after"],
            "payload": {
                # Include the previously-deployed SHA, so we ensure that the deployment
                # can create the appropriate commands, diffing the active deployment
                # against the requested one
                "previous_commit": previous_commit
            },
        },
    )

    if response.status_code != 201:
        print("Failed to create deployment")
        print(response.json())
        db.update_deployment_status(
            Log(
                form["after"],
                "Failed to create deployment",
                revision2=previous_commit,
                status="failure",
            )
        )
    else:
        db.add_deployment_log(
            Log(
                form["after"],
                "Deployment created",
                revision2=previous_commit,
                status="success",
            )
        )

    return response.status_code == 201


def is_against_primary_branch(deploy_config: dict, ref: str) -> bool:
    """
    Checks if the push is against the primary branch
    """
    return ref.endswith("/" + deploy_config["git"]["primary-branch"])
