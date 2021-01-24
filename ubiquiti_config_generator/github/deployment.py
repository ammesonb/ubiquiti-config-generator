"""
The deployment of configurations
"""
import traceback

from ubiquiti_config_generator.github import api, deploy_helper


def handle_deployment(form: dict, deploy_config: dict, access_token: str) -> bool:
    """
    Handles actual deploying of configuration
    """
    before = form["payload"]["previous_commit"]
    after = form["ref"]

    # TODO: mark deploy as in process

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

    try:
        router_connection = deploy_helper.get_router_connection(deploy_config)
    # Do want to catch everything here, to ensure the deployment status is updated
    # pylint: disable=broad-except
    except Exception as exception:
        # TODO: update deployment status here
        return False

    shell_file_base = (
        f"{deploy_config['router']['command-file-path']}/{before}..{after}-[###].sh"
    )

    file_names = []

    group_index = 0
    for commands in command_groups:
        # For now, while figuring out what to do
        # pylint: disable=unused-variable
        file_name = shell_file_base.replace("[###]", str(group_index).rjust(3, "0"))
        file_names.append(file_name)
        bash_contents = deploy_helper.generate_bash_commands(commands, deploy_config)
        if not deploy_helper.write_data_to_router_file(
            router_connection, file_name, bash_contents
        ):
            raise ValueError(f"Failed to write {file_name} to router!")

    # TODO: make one giant bash file to execute all the pieces
    mega_file = ""

    for file_name in file_names:
        mega_file += f"$(which vbash) {file_name} $$\n"

    mega_file += "\nexit 0\n"

    if not deploy_helper.write_data_to_router_file(
        router_connection, shell_file_base.replace("-[###]", ""), mega_file
    ):
        raise ValueError("Failed to create aggregated command file on router!")

    return True
