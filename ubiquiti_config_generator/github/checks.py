#!/usr/bin/env python3
"""
Contains the interactions for GitHub webhooks
"""
from datetime import datetime, timezone
import time

from ubiquiti_config_generator import root_parser
from ubiquiti_config_generator.github import deploy_helper, api
from ubiquiti_config_generator.github.api import GREEN_CHECK, RED_CROSS


def handle_check_suite(form: dict, access_token: str) -> None:
    """
    Performs needed actions if a check suite event occurs
    """
    if form["action"] not in ["requested", "rerequested"]:
        print("Ignoring check_suite action " + form["action"])
        return

    head_sha = form.get("check_suite", form.get("check_run")).get("head_sha")

    print("Requesting a check")

    response = api.send_github_request(
        form["repository"]["url"] + "/check-runs",
        "post",
        access_token,
        {"name": "configuration-validator", "head_sha": head_sha,},
    )

    if response.status_code != 201:
        api.set_commit_status(
            form["repository"]["statuses_url"],
            head_sha,
            access_token,
            "failure",
            "Could not schedule check run",
        )
        print(f"Failed to schedule check: got status {response.status_code}!")
        print(response.json())
    else:
        print("Check requested successfully")


# This probably should be refactored, but error handling here seems appropriate
# which causes a need to break this frequently - consider redesign at some point
# pylint: disable=too-many-return-statements
def process_check_run(deploy_config: dict, form: dict, access_token: str) -> bool:
    """
    Processes and completes a check run
    """
    if form["action"] not in ["created", "rerequested", "requested_action"]:
        print("Ignoring check_run action " + form["action"])
        return True

    if not api.update_check(
        access_token,
        form["check_run"]["url"],
        "in_progress",
        # Can't mock utcnow(), so use time.time instead
        {"started_at": datetime.fromtimestamp(time.time(), timezone.utc).isoformat()},
    ):
        return False

    try:
        deployed_sha = api.get_active_deployment_sha(
            form["repository"]["deployments_url"], access_token
        )
    except ValueError:
        finalize_check_state(["Failed to get deployed commit"], form, access_token)
        return False

    api.setup_config_repo(
        access_token,
        [
            api.Repository(
                deploy_config["git"]["config-folder"],
                form["repository"]["clone_url"],
                deployed_sha,
            ),
            api.Repository(
                deploy_config["git"]["diff-config-folder"],
                form["repository"]["clone_url"],
                form["check_run"]["head_sha"],
            ),
        ],
    )

    try:
        production_config_node = root_parser.RootNode.create_from_configs(
            deploy_config["git"]["config-folder"]
        )
        branch_config_node = root_parser.RootNode.create_from_configs(
            deploy_config["git"]["diff-config-folder"]
        )
    # No exception should occur here - fail if anything goes wrong
    # pylint: disable=broad-except
    except Exception as exception:
        api.update_check_with_exception(
            "Failed to load configuration",
            access_token,
            form["check_run"]["url"],
            exception,
        )
        finalize_commit_status(
            form["repository"]["statuses_url"],
            form["check_run"]["head_sha"],
            access_token,
            "failure",
        )
        return False

    try:
        branch_config_node.validate()
    # No exception should occur here - fail if anything goes wrong
    # pylint: disable=broad-except
    except Exception as exception:
        api.update_check_with_exception(
            "Exception occurred during validation",
            access_token,
            form["check_run"]["url"],
            exception,
        )
        finalize_commit_status(
            form["repository"]["statuses_url"],
            form["check_run"]["head_sha"],
            access_token,
            "failure",
        )
        return False

    if not finalize_check_state(
        branch_config_node.validation_failures(), form, access_token
    ):
        return False

    success = True
    if form["check_run"]["pull_requests"]:
        comment = get_pr_comment(
            deploy_config, branch_config_node, production_config_node
        )
        for pull in form["check_run"]["pull_requests"]:
            if not api.add_comment(access_token, pull["url"], comment):
                success = False

    return success


def get_output_of_validations(validations: list) -> dict:
    """
    Creates an output summary for the validation check
    """
    output = {
        # Can't mock utcnow(), so use time.time instead
        "completed_at": datetime.fromtimestamp(time.time(), timezone.utc).isoformat()
    }
    if validations:
        output.update(
            {
                "conclusion": "failure",
                "output": {
                    "title": "Configuration Validator",
                    "summary": f"{RED_CROSS} Invalid configuration",
                    "text": "- " + "\n- ".join(validations),
                },
            }
        )
    else:
        output.update(
            {
                "conclusion": "success",
                "output": {
                    "title": "Configuration Validator",
                    "summary": f"{GREEN_CHECK} Configuration successfully validated",
                },
            }
        )

    return output


def finalize_check_state(validations: list, form: dict, access_token: str) -> bool:
    """
    Set the results of the check
    """
    output = get_output_of_validations(validations)
    commit_status_set = finalize_commit_status(
        form["repository"]["statuses_url"],
        form["check_run"]["head_sha"],
        access_token,
        output["conclusion"],
    )

    check_updated = api.update_check(
        access_token, form["check_run"]["url"], "completed", output
    )
    return commit_status_set and check_updated


def finalize_commit_status(
    statuses_url: str, sha: str, access_token: str, conclusion: str
) -> bool:
    """
    In theory, this should happen automatically when the check status is updated
    BUT, the status API returns nothing and we do want to block deployments for
    pending and failed pushes, so may as well update the status manually,
    just to be safe
    """
    return api.set_commit_status(
        statuses_url.replace("{/sha}", "/" + sha), sha, access_token, conclusion,
    )


def get_pr_comment(
    deploy_config: dict,
    branch_config_node: root_parser.RootNode,
    production_config_node: root_parser.RootNode,
) -> str:
    """
    Gets the comment to add to the PR for the current configurations
    """
    comment = api.summarize_deploy_config_choices(deploy_config)
    comment += "\n"

    differences = deploy_helper.diff_configurations(
        branch_config_node.get_commands()[1], production_config_node.get_commands()[1],
    )
    for category in ["added", "removed", "changed"]:
        commands = getattr(differences, category)
        if not commands:
            continue

        comment += f"## Commands {category}:\n\n- " + "\n- ".join(
            [command + " " + str(commands[command]) for command in commands]
        )

        comment += "\n"

    return comment.strip()
