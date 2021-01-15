#!/usr/bin/env python3
"""
Contains the interactions for GitHub webhooks
"""
from datetime import datetime, timezone
import hashlib
import hmac
import os
from os import path
import shutil
import subprocess
import time
import traceback

import jwt
import requests

from ubiquiti_config_generator import root_parser
from ubiquiti_config_generator import deploy_helper

GH_HEADER = {
    "accept": "application/vnd.github.v3+json",
}

GH_JWT_HEADER = lambda jwt: {**GH_HEADER, "Authorization": "Bearer " + jwt}
GH_TOKEN_HEADER = lambda token: {**GH_HEADER, "Authorization": "token " + token}

GREEN_CHECK = "&#9989;"
RED_CROSS = "&#10060;"
WARNING = "&#x26a0;"


def get_jwt(deploy_config: dict) -> str:
    """
    Get the JSON web token, needed for basic API access
    """
    return jwt.encode(
        {
            # Request starting now
            "iat": int(time.time()),
            # Token expires in ten minutes
            "exp": int(time.time()) + 600,
            "iss": deploy_config["git"]["app-id"],
        },
        open(deploy_config["git"]["private-key-path"], "rb").read(),
        "RS256",
    )


def get_access_token(jwt_token: str) -> str:
    """
    Gets an access token, used for fully-authenticated app actions
    """
    installations = requests.get(
        "https://api.github.com/app/installations", headers=GH_JWT_HEADER(jwt_token)
    )

    response = requests.post(
        installations.json()[0]["access_tokens_url"], headers=GH_JWT_HEADER(jwt_token)
    )
    return response.json()["token"]


def validate_message(deploy_config: dict, body: str, sha: str) -> bool:
    """
    Check if the request body is valid
    """
    signature = hmac.new(
        deploy_config["git"]["webhook-secret"].encode(), body, hashlib.sha256
    ).hexdigest()
    return sha == "sha256=" + signature


def handle_check_suite(form: dict, access_token: str) -> None:
    """
    Performs needed actions if a check suite event occurs
    """
    if form["action"] not in ["requested", "rerequested"]:
        print("Ignoring check_suite action " + form["action"])
        return

    head_sha = form.get("check_suite", form.get("check_run")).get("head_sha")

    print("Requesting a check")

    response = requests.post(
        form["repository"]["url"] + "/check-runs",
        json={"name": "configuration-validator", "head_sha": head_sha,},
        headers=GH_TOKEN_HEADER(access_token),
    )

    if response.status_code != 201:
        print(f"Failed to schedule check: got status {response.status_code}!")
        print(response.json())
    else:
        print("Check requested successfully")


def process_check_run(deploy_config: dict, form: dict, access_token: str) -> None:
    """
    Processes and completes a check run
    """
    if form["action"] not in ["created", "rerequested", "requested_action"]:
        print("Ignoring check_run action " + form["action"])
        return

    if not update_check(
        access_token,
        form["check_run"]["url"],
        "in_progress",
        # Can't mock utcnow(), so use time.time instead
        {"started_at": datetime.fromtimestamp(time.time(), timezone.utc).isoformat()},
    ):
        return

    for folder in [
        deploy_config["git"]["config-folder"],
        deploy_config["git"]["diff-config-folder"],
    ]:
        clone_repository(
            access_token, "github.com/" + form["repository"]["full_name"], folder,
        )

    checkout(deploy_config["git"]["diff-config-folder"], form["check_run"]["head_sha"])

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
        update_check_with_exception(
            "Failed to load configuration",
            access_token,
            form["check_run"]["url"],
            exception,
        )
        return

    try:
        valid = branch_config_node.validate()
    # No exception should occur here - fail if anything goes wrong
    # pylint: disable=broad-except
    except Exception as exception:
        update_check_with_exception(
            "Exception occurred during validation",
            access_token,
            form["check_run"]["url"],
            exception,
        )
        return

    output = {
        # Can't mock utcnow(), so use time.time instead
        "completed_at": datetime.fromtimestamp(time.time(), timezone.utc).isoformat()
    }
    if not valid:
        output.update(
            {
                "conclusion": "failure",
                "output": {
                    "title": "Configuration Validator",
                    "summary": f"{RED_CROSS} Invalid configuration",
                    "text": "- "
                    + "\n- ".join(branch_config_node.validation_failures()),
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

    if not update_check(access_token, form["check_run"]["url"], "completed", output):
        return

    if form["check_run"]["pull_requests"]:
        comment = summarize_deploy_config_choices(deploy_config)
        comment += "\n"

        differences = deploy_helper.diff_configurations(
            branch_config_node.get_commands()[1],
            production_config_node.get_commands()[1],
        )
        for category in ["added", "removed", "changed"]:
            commands = getattr(differences, category)
            if not commands:
                continue

            comment += f"## Commands {category}:\n\n- " + "\n- ".join(
                [command + " " + commands[command] for command in commands]
            )

            comment += "\n"

        for pull in form["check_run"]["pull_requests"]:
            add_comment(access_token, pull["url"], comment)


def update_check_with_exception(
    error_text: str, access_token: str, check_url: str, exception: Exception
) -> None:
    """
    Update check indicating an exception as occurred
    """
    print("Caught exception " + str(exception) + " during check!")
    output = {
        # Can't mock utcnow(), so use time.time instead
        "completed_at": datetime.fromtimestamp(time.time(), timezone.utc).isoformat(),
        "conclusion": "failure",
        "output": {
            "title": "Configuration Validator",
            "summary": f"{RED_CROSS} {error_text}",
            "text": str(exception) + "\n\n" + traceback.format_exc(),
        },
    }
    update_check(access_token, check_url, "completed", output)


def update_check(
    access_token: str, check_url: str, status: str, extra_data: dict
) -> bool:
    """
    Updates a given check with a status and data
    """
    check_id = check_url.split("/")[-1]
    print(f"Updating check {check_id} to {status}")
    response = requests.patch(
        check_url,
        json={"status": status, **extra_data},
        headers=GH_TOKEN_HEADER(access_token),
    )
    if response.status_code != 200:
        print(f"Check update {check_id} failed with code {response.status_code}")
        print(response.json())

    return response.status_code == 200


def add_comment(access_token: str, pull_url: str, comment: str):
    """
    Add a comment to a PR
    """
    pull = requests.get(pull_url, headers=GH_TOKEN_HEADER(access_token))
    if pull.status_code != 200:
        print(f"Failed to get pull request {pull_url}")
        return

    print("Posting comment")
    posted = requests.post(
        pull.json()["comments_url"],
        json={"body": comment},
        headers=GH_TOKEN_HEADER(access_token),
    )
    if posted.status_code != 201:
        print(f"Failed to post comment to review {pull_url}")
        print(posted.json())


def clone_repository(
    access_token: str, repo: str, clone_path: str, remove_existing_folder: bool = True
) -> None:
    """
    Clones a repository to a given folder, optionally removing existing
    """
    if remove_existing_folder and path.exists(clone_path):
        print(f"Deleting files in {clone_path}")
        shutil.rmtree(clone_path)

    print("Cloning {0} into {1}".format("/".split(repo)[-1], clone_path))
    subprocess.run(
        [
            "git",
            "clone",
            "--quiet",
            f"https://x-access-token:{access_token}@{repo}",
            clone_path,
        ],
        check=True,
    )


def checkout(repo_path: str, sha_or_branch: str) -> None:
    """
    Check out a SHA or branch in a repo
    """
    print(f"Checking out {sha_or_branch} in {repo_path}")
    cwd = os.getcwd()
    os.chdir(repo_path)

    try:
        subprocess.run(["git", "checkout", "--quiet", sha_or_branch], check=True)
    finally:
        os.chdir(cwd)


def summarize_deploy_config_choices(deploy_config: dict) -> str:
    """
    Creates a summary of the deploy configuration, for displaying
    on a GitHub comment to ensure it matches what you want to do
    """
    difference_only = (
        GREEN_CHECK
        + " "
        + (
            "Applying *DIFFERENCE* only"
            if deploy_config["apply-difference-only"]
            else "Applying *ENTIRE* configuration"
        )
    )
    rollback = (
        f"{GREEN_CHECK} Will rollback on configuration error"
        if deploy_config["auto-rollback-on-failure"]
        else f"{WARNING} Will **NOT** rollback on configuration error"
    )
    auto_restart = (
        f"{GREEN_CHECK} Will restart after {deploy_config['reboot-after-minutes']} "
        "minutes without confirm"
        if deploy_config["reboot-after-minutes"]
        else f"{WARNING} Will **NOT** automatically restart"
    )
    auto_save = (
        f"{GREEN_CHECK} Will **NOT** save automatically"
        if not deploy_config["save-after-commit"]
        else f"{WARNING} **WILL** save configuration immediately after commit"
    )
    return "\n".join(
        [
            "## Deployment overview",
            "- " + difference_only,
            "- " + rollback,
            "- " + auto_restart,
            "- " + auto_save,
        ]
    )
