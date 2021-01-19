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
        print(f"Failed to schedule check: got status {response.status_code}!")
        print(response.json())
    else:
        print("Check requested successfully")


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

    api.setup_config_repo(
        access_token,
        "github.com/" + form["repository"]["full_name"],
        deploy_config,
        form["check_run"]["head_sha"],
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
        return False

    output = get_output_of_validations(branch_config_node.validation_failures())

    # In theory, this should happen automatically when the check status is updated
    # BUT, the status API returns nothing and we do want to block deployments for
    # pending and failed pushes, so may as welll, just to be safe
    if not api.set_commit_status(
        form["repository"]["statuses_url"].replace(
            "{/sha}", "/" + form["check_run"]["head_sha"]
        ),
        form["check_run"]["head_sha"],
        access_token,
        output["conclusion"],
    ):
        return False

    if not api.update_check(
        access_token, form["check_run"]["url"], "completed", output
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
