"""
Interactions for pushes from GitHub, triggering the deploy process
"""
from typing import Optional

from ubiquiti_config_generator.github import api


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

    response = api.send_github_request(
        form["repository"]["deployments_url"],
        "post",
        access_token,
        {
            "ref": form["after"],
            "payload": {
                # Get the previously-deployed SHA, so we ensure that the deployment
                # can create the appropriate commands, diffing the active deployment
                # against the requested one
                "previous_commit": api.get_active_deployment_sha(
                    form["repository"]["deployments_url"], access_token
                )
            },
        },
    )

    # TODO: is 201 correct here?
    if response.status_code != 201:
        print("Failed to create deployment!")
        print(response.json())
        return False

    return response.status_code == 201


def is_against_primary_branch(deploy_config: dict, ref: str) -> bool:
    """
    Checks if the push is against the primary branch
    """
    return ref.endswith("/" + deploy_config["git"]["primary-branch"])
