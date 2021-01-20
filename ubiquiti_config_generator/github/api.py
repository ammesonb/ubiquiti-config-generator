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


def setup_config_repo(
    access_token: str, repository_url: str, deploy_config: dict, previous_sha: str = ""
):
    """
    Set up the configuration directory
    """
    repos = [deploy_config["git"]["config-folder"]]
    if previous_sha:
        repos.append(deploy_config["git"]["diff-config-folder"])

    for folder in repos:
        clone_repository(
            access_token, repository_url, folder,
        )

    if previous_sha:
        checkout(repos[1], previous_sha)


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


def set_commit_status(url: str, sha: str, access_token: str, status: str) -> bool:
    """
    Sets the status of a commit
    It appears that checks also set commit statuses implicitly, but the API does not
    return those statuses as directly attached to the commit....
    There is no documentation I can find indicating this one way or the other
    """
    response = send_github_request(
        url,
        "post",
        access_token,
        {
            "state": status,
            "description": "Configuration validation",
            "context": "configuration-validator",
        },
    )
    if response.status_code != 201:
        print(f"Failed to set commit status {status} for {sha}")
        print(response.json())

    return response.status_code == 201


def add_comment(access_token: str, pull_url: str, comment: str) -> bool:
    """
    Add a comment to a PR
    """
    pull = requests.get(pull_url, headers=GH_TOKEN_HEADER(access_token))
    if pull.status_code != 200:
        print(f"Failed to get pull request {pull_url}")
        print(pull.json())
        return False

    print("Posting comment")
    posted = requests.post(
        pull.json()["comments_url"],
        json={"body": comment},
        headers=GH_TOKEN_HEADER(access_token),
    )
    if posted.status_code != 201:
        print(f"Failed to post comment to review {pull_url}")
        print(posted.json())
        return False

    return True


def clone_repository(
    access_token: str, repo: str, clone_path: str, remove_existing_folder: bool = True
) -> None:
    """
    Clones a repository to a given folder, optionally removing existing
    """
    if remove_existing_folder and path.exists(clone_path):
        print(f"Deleting files in {clone_path}")
        shutil.rmtree(clone_path)

    print("Cloning {0} into {1}".format(repo.split("/")[-1], clone_path))
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


def send_github_request(
    url: str, method: str, access_token: str, json_data: dict = None
) -> requests.Response:
    """
    Send a request to github
    """
    return (
        getattr(requests, method)(
            url, json=json_data, headers=GH_TOKEN_HEADER(access_token)
        )
        if json_data
        else getattr(requests, method)(url, headers=GH_TOKEN_HEADER(access_token))
    )