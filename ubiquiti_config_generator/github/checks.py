#!/usr/bin/env python3
"""
Contains the interactions for GitHub webhooks
"""
from datetime import datetime, timezone
import time

from ubiquiti_config_generator import root_parser
from ubiquiti_config_generator.github import deploy_helper, api
from ubiquiti_config_generator.github.api import GREEN_CHECK, RED_CROSS
from ubiquiti_config_generator.messages import db
from ubiquiti_config_generator.messages.check import Check
from ubiquiti_config_generator.messages.log import Log


def handle_check_suite(form: dict, access_token: str) -> None:
    """
    Performs needed actions if a check suite event occurs
    """
    if form["action"] not in ["requested", "rerequested"]:
        print("Ignoring check_suite action " + form["action"])
        return

    head_sha = form.get("check_suite", form.get("check_run")).get("head_sha")
    db.create_check(Check(head_sha, "pending", time.time()))

    print("Requesting a check")

    response = api.send_github_request(
        form["repository"]["url"] + "/check-runs",
        "post",
        access_token,
        {
            "name": "configuration-validator",
            "head_sha": head_sha,
            "details_url": form["check_suite"]["app"]["external_url"]
            + f"/checks/{head_sha}",
        },
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

        db.update_check_status(
            Log(head_sha, "Failed to create a check run", status="failure")
        )
        print(response.json())
    else:
        print("Check requested successfully")
        db.add_check_log(Log(head_sha, "Check run scheduled"))


# This probably should be refactored, but error handling here seems appropriate
# which causes a need to break this frequently - consider redesign at some point
# pylint: disable=too-many-return-statements
def process_check_run(deploy_config: dict, form: dict, access_token: str) -> bool:
    """
    Processes and completes a check run
    """
    if form["action"] not in ["created", "rerequested", "requested_action"]:
        print("Ignoring check_run action " + form["action"])
        db.update_check_status(
            Log(
                form["check_run"]["head_sha"],
                f"Action {form['action']} does not require processing",
                status="success",
            )
        )
        return True

    if not api.update_check(
        access_token,
        form["check_run"]["url"],
        "in_progress",
        # Can't mock utcnow(), so use time.time instead
        {"started_at": datetime.fromtimestamp(time.time(), timezone.utc).isoformat()},
    ):
        db.update_check_status(
            Log(
                form["check_run"]["head_sha"],
                "Failed to update check to in progress",
                status="failure",
            )
        )
        return False

    db.add_check_log(
        Log(
            form["check_run"]["head_sha"],
            "Fetching latest successful revision from Git history",
        )
    )

    try:
        deployed_sha = api.get_active_deployment_sha(
            form["repository"]["deployments_url"], access_token
        )
    except ValueError:
        finalize_check_state(["Failed to get deployed commit"], form, access_token)
        db.update_check_status(
            Log(
                form["check_run"]["head_sha"],
                "Failed to get deployed commit",
                status="failure",
            )
        )
        return False

    db.add_check_log(Log(form["check_run"]["head_sha"], "Cloning repositories",))

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

    db.add_check_log(
        Log(
            form["check_run"]["head_sha"], "Loading current and feature configurations",
        )
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
        db.update_check_status(
            Log(
                form["check_run"]["head_sha"],
                "Failed to load configuration",
                status="failure",
            )
        )
        return False

    db.add_check_log(
        Log(form["check_run"]["head_sha"], "Validating feature configuration",)
    )

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
        db.update_check_status(
            Log(
                form["check_run"]["head_sha"],
                "Exception occurred during validation",
                status="failure",
            )
        )
        return False

    db.add_check_log(
        Log(form["check_run"]["head_sha"], "Reporting final configuration state",)
    )

    validation_failures = branch_config_node.validation_failures()
    if not finalize_check_state(validation_failures, form, access_token):
        db.update_check_status(
            Log(
                form["check_run"]["head_sha"],
                "Failed to update check state on GitHub",
                status="failure",
            )
        )
        return False

    db.add_check_log(
        Log(form["check_run"]["head_sha"], "Updating PR/s with deployment details",)
    )

    success = True
    if form["check_run"]["pull_requests"]:
        comment = get_pr_comment(
            deploy_config, branch_config_node, production_config_node
        )
        for pull in form["check_run"]["pull_requests"]:
            if not api.add_comment(access_token, pull["url"], comment):
                success = False

    if validation_failures:
        db.update_check_status(
            Log(form["check_run"]["head_sha"], "Check failed!", status="failure")
        )

        for failure in validation_failures:
            db.add_check_log(Log(form["check_run"]["head_sha"], failure))
    else:
        db.update_check_status(
            Log(form["check_run"]["head_sha"], "Check complete", status="success")
        )

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

        comment += f"## Commands {category}:\n\n"
        for command_key, command_value in commands.items():
            if isinstance(command_value, list):
                for each_value in command_value:
                    comment += f"- {command_key} {each_value}\n"
            else:
                comment += f"- {command_key} {command_value}\n"

    return comment.strip()
