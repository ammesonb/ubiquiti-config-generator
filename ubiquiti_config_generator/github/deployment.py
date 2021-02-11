"""
The deployment of configurations
"""
from typing import List

import paramiko

from ubiquiti_config_generator.github import api, deploy_helper
from ubiquiti_config_generator.github.deployment_metadata import DeployMetadata
from ubiquiti_config_generator.messages import db
from ubiquiti_config_generator.messages.log import Log


def handle_deployment(form: dict, deploy_config: dict, access_token: str) -> bool:
    """
    Handles actual deploying of configuration
    """
    if form["action"] != "created":
        print(f"Ignoring deployment action {form['action']}")
        return True

    before = form["payload"]["previous_commit"]
    after = form["ref"]

    metadata = DeployMetadata(
        before,
        after,
        deploy_config["git"]["webhook-url"],
        form["deployment"]["statuses_url"],
        deploy_config,
    )

    if not db.update_deployment_status(
        Log(before, "Deploy in progress", revision2=after, status="in_progress")
    ):
        print("Failed to update local copy of deploy to in progress")

    api.update_deployment_state(
        metadata.status_url,
        metadata.external_app_url,
        before,
        after,
        access_token,
        "in_progress",
    )

    api.setup_config_repo(
        access_token,
        [
            api.Repository(
                deploy_config["git"]["config-folder"],
                form["repository"]["clone_url"],
                after,
            ),
            api.Repository(
                deploy_config["git"]["diff-config-folder"],
                form["repository"]["clone_url"],
                before,
            ),
        ],
    )

    command_groups = deploy_helper.get_commands_to_run(
        deploy_config["git"]["diff-config-folder"],
        deploy_config["git"]["config-folder"],
        deploy_config["apply-difference-only"],
    )

    result = (
        "success"
        if load_and_execute_config_changes(command_groups, metadata, access_token)
        else "failure"
    )

    if not db.update_deployment_status(
        Log(before, "Deploy completed", revision2=after, status=result)
    ):
        print(f"Failed to update local copy of deploy to {result}")

    api.update_deployment_state(
        metadata.status_url,
        metadata.external_app_url,
        before,
        after,
        access_token,
        result,
    )

    return result == "success"


def load_and_execute_config_changes(
    command_groups: List[List[str]], metadata: DeployMetadata, access_token: str
) -> bool:
    """
    Load the configuration onto the router and execute it
    """
    try:
        db.add_deployment_log(
            Log(
                metadata.before_sha,
                "Connecting to router",
                revision2=metadata.after_sha,
            )
        )
        router_connection = deploy_helper.get_router_connection(
            metadata.deployment_configuration
        )
        # Pylint doesn't detect usage in f strings I guess
        # pylint: disable=unused-variable
        config_deploy_file = send_config_files_to_router(
            router_connection, metadata, command_groups
        )
        db.add_deployment_log(
            Log(
                metadata.before_sha,
                "Command files deployed to the router",
                revision2=metadata.after_sha,
            )
        )
        output = router_connection.exec_command("bash {config_deploy_file}")

    except ValueError as error:
        fail_deployment(
            metadata.before_sha,
            metadata.after_sha,
            metadata.status_url,
            metadata.external_app_url,
            access_token,
            "Failed to prepare for deployment - got:\n" + str(error),
        )
        return False

    except paramiko.SSHException as error:
        fail_deployment(
            metadata.before_sha,
            metadata.after_sha,
            metadata.status_url,
            metadata.external_app_url,
            access_token,
            "Failed to execute deployment - got:\n" + str(error),
        )

        return False

    # We do want to catch everything else here,
    # to ensure the deployment status is updated
    # pylint: disable=broad-except
    except Exception as error:
        fail_deployment(
            metadata.before_sha,
            metadata.after_sha,
            metadata.status_url,
            metadata.external_app_url,
            access_token,
            str(error),
        )
        return False

    log_command_output(metadata.before_sha, metadata.after_sha, output)
    return True


def send_config_files_to_router(
    router_connection: paramiko.Channel,
    metadata: DeployMetadata,
    command_groups: List[List[str]],
) -> str:
    """
    Creates the remote bash scripts to run on the router
    Returns the name of the aggregate script to execute
    """
    shell_file_base = (
        f"{metadata.deployment_configuration['router']['command-file-path']}/"
        f"{metadata.before_sha}..{metadata.after_sha}-[###].sh"
    )

    file_names = []
    group_index = 0
    for commands in command_groups:
        # For now, while figuring out what to do
        # pylint: disable=unused-variable
        file_name = shell_file_base.replace("[###]", str(group_index).rjust(3, "0"))
        file_names.append(file_name)

        # pylint: disable=too-many-function-args
        send_file_to_router(
            metadata.before_sha,
            metadata.after_sha,
            router_connection,
            group_index,
            commands,
            metadata.deployment_configuration,
            file_name,
        )

        group_index += 1

    aggregate_file_name = shell_file_base.replace("-[###]", "")
    send_aggregate_file_to_router(
        metadata.before_sha,
        metadata.after_sha,
        router_connection,
        file_names,
        aggregate_file_name,
    )

    return aggregate_file_name


# pylint: disable=too-many-arguments
def send_file_to_router(
    before: str,
    after: str,
    router_connection: paramiko.SSHClient,
    commands: List[str],
    deploy_config: dict,
    file_name: str,
):
    """
    Sends a set of commands to the router
    """
    bash_contents = deploy_helper.generate_bash_commands(commands, deploy_config)
    db.add_deployment_log(
        Log(before, f"Adding command set {file_name} to router", revision2=after)
    )
    if not deploy_helper.write_data_to_router_file(
        router_connection, file_name, bash_contents
    ):
        print(f"Failed to write {file_name} to router")
        raise ValueError("Failed to write file to router")


def send_aggregate_file_to_router(
    before: str,
    after: str,
    router_connection: paramiko.SSHClient,
    file_names: List[str],
    aggregate_file_name: str,
):
    """
    Sends the aggregate commands file to the router
    """
    mega_file = ""

    for file_name in file_names:
        mega_file += f"$(which vbash) {file_name} $$\n"

    mega_file += "\nexit 0\n"

    db.add_deployment_log(
        Log(
            before,
            f"Adding combined command set {aggregate_file_name} to router",
            revision2=after,
        )
    )
    if not deploy_helper.write_data_to_router_file(
        router_connection, aggregate_file_name, mega_file
    ):
        print("Failed to create aggregated command file on router")
        raise ValueError("Failed to create aggregated command file on router")


# pylint: disable=too-many-arguments
def fail_deployment(
    before: str,
    after: str,
    deployment_url: str,
    external_app_url: str,
    access_token: str,
    failure_reason: str,
):
    """
    Fails the deployment
    """
    api.update_deployment_state(
        deployment_url,
        external_app_url,
        before,
        after,
        access_token,
        "failure",
        "Failed to run configuration deployment",
    )

    if not db.add_deployment_log(
        Log(before, failure_reason, revision2=after, status="failure",)
    ):
        print("Failed to record deployment execution failure in DB")


def log_command_output(before: str, after: str, output: tuple):
    """
    .
    """
    _, stdout, stderr = output

    for log in stdout.split("\n") + stderr.split("\n"):
        if not log:
            continue
        if not db.add_deployment_log(Log(before, log, revision2=after,)):
            print("Failed to add deployment execution log in DB")
