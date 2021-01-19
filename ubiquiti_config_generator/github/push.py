"""
Interactions for pushes from GitHub, triggering the deploy process
"""
from typing import Optional

from ubiquiti_config_generator.github import api


def is_against_primary_branch(deploy_config: dict, ref: str) -> bool:
    """
    Checks if the push is against the primary branch
    """
    return ref.endswith("/" + deploy_config["git"]["primary-branch"])


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

    # Currently naively assuming previous deploy worked
    # Should get currently-active deployment by:
    # - List all deployments and their statuses (separate call), start with most recent
    # - If deploy fails, mark as failed on GitHub, that way it won't
    #   screw up the next deployment attempt, and we preserve the difference
    #   in configuration commands if only applying a difference between live and
    #   modified configuration
    api.send_github_request(
        form["repository"]["deployments_url"],
        "post",
        access_token,
        {"ref": form["after"], "payload": {"previous_commit": form["before"]}},
    )
