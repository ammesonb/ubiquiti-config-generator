"""
The deployment of configurations
"""

from ubiquiti_config_generator.github import api, deploy_helper


def handle_new_deployment(form: dict, deploy_config: dict, access_token: str):
    """
    Handles actual deploying of configuration
    """
    before = form["payload"]["previous_commit"]
    after = form["ref"]

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

    shell_file_base = f"{before}..{after}-[###].sh"
    group_index = 0
    for commands in command_groups:
        # For now, while figuring out what to do
        # pylint: disable=unused-variable
        file_name = shell_file_base.replace("[###]", str(group_index).rjust(3, "0"))
        bash_contents = deploy_helper.generate_bash_commands(commands, deploy_config)
        # TODO: how to get this to the router?
