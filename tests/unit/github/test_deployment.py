"""
Tests github deployment code
"""
import paramiko
import pytest

from ubiquiti_config_generator.github import api, deployment, deploy_helper
from ubiquiti_config_generator.github.deployment_metadata import DeployMetadata
from ubiquiti_config_generator.messages import db
from ubiquiti_config_generator.messages.log import Log
from ubiquiti_config_generator.testing_utils import counter_wrapper


def test_log_command_output(monkeypatch, capsys):
    """
    .
    """
    messages = [
        "a message",
        "a failure",
        "details about failure",
        "more details about failure",
    ]

    # pylint: disable=unused-argument
    @counter_wrapper
    def add_log(log: Log, cursor=None) -> bool:
        """
        .
        """
        assert (
            log.message == messages[add_log.counter - 1]
        ), f"Message {add_log.counter} logged is correct"
        return add_log.counter < 4

    monkeypatch.setattr(db, "add_deployment_log", add_log)

    deployment.log_command_output(
        "beforesha",
        "aftersha",
        (
            "",
            "a message\na failure\n",
            "details about failure\nmore details about failure\n",
        ),
    )
    assert add_log.counter == 4, "Four logs added"
    printed = capsys.readouterr()
    assert (
        printed.out == "Failed to add deployment execution log in DB\n"
    ), "Single DB failure printed"


def test_fail_deployment(monkeypatch, capsys):
    """
    .
    """
    # pylint: disable=unused-argument
    @counter_wrapper
    def patch_update_deployment(*args, **kwargs):
        """
        .
        """

    monkeypatch.setattr(api, "update_deployment_state", patch_update_deployment)
    monkeypatch.setattr(db, "add_deployment_log", lambda *args, **kwargs: False)

    deployment.fail_deployment("before", "after", "/deploys", "/app", "abc123", "error")
    assert patch_update_deployment.counter == 1, "Deployment state updated"
    printed = capsys.readouterr()
    assert (
        printed.out == "Failed to record deployment execution failure in DB\n"
    ), "Output printed"

    monkeypatch.setattr(db, "add_deployment_log", lambda *args, **kwargs: True)
    deployment.fail_deployment("before", "after", "/deploys", "/app", "abc123", "error")
    assert patch_update_deployment.counter == 2, "Deployment state updated"
    printed = capsys.readouterr()
    assert printed.out == "", "Nothing printed"


def test_send_file_to_router(monkeypatch, capsys):
    """
    .
    """
    # pylint: disable=unused-argument
    @counter_wrapper
    def get_commands(*args, **kwargs):
        """
        .
        """
        return "commands"

    @counter_wrapper
    def add_log(log: Log, cursor=None):
        """
        .
        """
        assert (
            log.message == "Adding command set /tmp/foo.sh to router"
        ), "Log has correct file name"

    monkeypatch.setattr(deploy_helper, "generate_bash_commands", get_commands)
    monkeypatch.setattr(db, "add_deployment_log", add_log)
    monkeypatch.setattr(
        deploy_helper, "write_data_to_router_file", lambda *args, **kwargs: False
    )

    with pytest.raises(ValueError):
        deployment.send_file_to_router(
            "before", "after", None, ["commands"], {}, "/tmp/foo.sh"
        )

    assert get_commands.counter == 1, "Commands generated"
    assert add_log.counter == 1, "Log added"
    printed = capsys.readouterr()
    assert printed.out == "Failed to write /tmp/foo.sh to router\n", "Error printed"

    monkeypatch.setattr(
        deploy_helper, "write_data_to_router_file", lambda *args, **kwargs: True
    )

    deployment.send_file_to_router(
        "before", "after", None, ["commands"], {}, "/tmp/foo.sh"
    )
    assert get_commands.counter == 2, "Commands generated"
    assert add_log.counter == 2, "Log added"


def test_send_aggregate_file_to_router(monkeypatch, capsys):
    """
    .
    """
    # pylint: disable=unused-argument
    @counter_wrapper
    def add_log(log: Log, cursor=None):
        """
        .
        """
        assert (
            log.message == "Adding combined command set /tmp/aggregate.sh to router"
        ), "Log message correct"

    @counter_wrapper
    def write_router_data(connection, file_name, file_data):
        """
        .
        """
        assert file_data == (
            "$(which vbash) /tmp/commands1.sh $$\n"
            "$(which vbash) /tmp/commands2.sh $$\n\n"
            "exit 0\n"
        ), "File data correct"
        return write_router_data.counter == 2

    monkeypatch.setattr(db, "add_deployment_log", add_log)
    monkeypatch.setattr(deploy_helper, "write_data_to_router_file", write_router_data)

    with pytest.raises(ValueError):
        deployment.send_aggregate_file_to_router(
            "before",
            "after",
            None,
            ["/tmp/commands1.sh", "/tmp/commands2.sh"],
            "/tmp/aggregate.sh",
        )

    assert add_log.counter == 1, "Deployment log added"
    assert write_router_data.counter == 1, "Attempted to write data to router"
    printed = capsys.readouterr()
    assert (
        printed.out == "Failed to create aggregated command file on router\n"
    ), "Error message printed"

    deployment.send_aggregate_file_to_router(
        "before",
        "after",
        None,
        ["/tmp/commands1.sh", "/tmp/commands2.sh"],
        "/tmp/aggregate.sh",
    )
    assert add_log.counter == 2, "Deployment log added"
    assert write_router_data.counter == 2, "Attempted to write data to router"
    printed = capsys.readouterr()
    assert printed.out == "", "No errors printed"


def test_send_config_files_to_router(monkeypatch):
    """
    .
    """

    # pylint: disable=unused-argument,too-many-arguments
    @counter_wrapper
    def send_file(
        before, after, router_connection, commands, deploy_config, file_name,
    ):
        """
        .
        """
        command_number = send_file.counter - 1
        assert commands == [f"command{command_number}"], "Command group correct"
        assert file_name == (
            "/tmp/bef123..afe321-0"
            + ("0" if command_number < 10 else "")
            + f"{command_number}.sh"
        ), "File name correct"

    @counter_wrapper
    def send_aggregate_file(
        before, after, router_connection, file_names, aggregate_file_name
    ):
        """
        .
        """
        assert aggregate_file_name == "/tmp/bef123..afe321.sh", "Aggregate file correct"

    monkeypatch.setattr(deployment, "send_file_to_router", send_file)
    monkeypatch.setattr(
        deployment, "send_aggregate_file_to_router", send_aggregate_file
    )

    assert (
        deployment.send_config_files_to_router(
            None,
            DeployMetadata(
                "bef123",
                "afe321",
                "",
                "/status",
                {"router": {"command-file-path": "/tmp"}},
            ),
            [
                ["command" + str(x)]
                # Send eleven, to ensure the padding is correct for multiple-digit numbers
                for x in range(11)
            ],
        )
        == "/tmp/bef123..afe321.sh"
    )

    assert send_file.counter == 11, "Eleven command files sent"
    assert send_aggregate_file.counter == 1, "Aggregate file sent"


def test_load_execute_config(monkeypatch):
    """
    .
    """
    # pylint: disable=unused-argument
    @counter_wrapper
    def add_log(*args, **kwargs):
        """
        .
        """

    @counter_wrapper
    def fail_router_connection(*args, **kwargs):
        """
        .
        """
        raise ValueError("Error connecting")

    @counter_wrapper
    def generic_fail(*args, **kwargs):
        """
        .
        """
        return AttributeError("Deploy config missing parameter or something")

    # pylint: disable=too-few-public-methods
    class FailRouter:
        """
        Fake router
        """

        # pylint: disable=no-self-use
        def exec_command(self, *args, **kwargs):
            """
            .
            """
            raise paramiko.SSHException("Failed to execute")

    # pylint: disable=too-few-public-methods
    class SuccessRouter:
        """
        Fake router
        """

        def exec_command(self, *args, **kwargs):
            """
            .
            """

    def get_fail_router(*args, **kwargs):
        """
        .
        """
        return FailRouter()

    def get_success_router(*args, **kwargs):
        """
        .
        """
        return SuccessRouter()

    @counter_wrapper
    def fail_send_config_files(*args, **kwargs):
        """
        .
        """
        raise ValueError("Error sending file")

    @counter_wrapper
    def fail_deployment(*args, **kwargs):
        """
        .
        """

    @counter_wrapper
    def log_output(*args, **kwargs):
        """
        .
        """

    monkeypatch.setattr(db, "add_deployment_log", generic_fail)
    monkeypatch.setattr(deployment, "fail_deployment", fail_deployment)

    assert not deployment.load_and_execute_config_changes(
        [], DeployMetadata("abc", "def", "/app", "/status", {}), "abc123"
    ), "Generic error  causes failure"

    assert generic_fail.counter == 1, "Tried to add deploy log but failed"
    assert fail_deployment.counter == 1, "Deployment failed"

    monkeypatch.setattr(db, "add_deployment_log", add_log)

    monkeypatch.setattr(deploy_helper, "get_router_connection", fail_router_connection)
    monkeypatch.setattr(
        deployment, "send_config_files_to_router", fail_send_config_files
    )
    monkeypatch.setattr(deployment, "fail_deployment", fail_deployment)
    monkeypatch.setattr(deployment, "log_command_output", log_output)

    assert not deployment.load_and_execute_config_changes(
        [], DeployMetadata("abc", "def", "/app", "/status", {}), "abc123"
    ), "Router conection fails"

    assert add_log.counter == 1, "Single log added"
    assert fail_send_config_files.counter == 0, "Did not try to send config files"
    assert fail_deployment.counter == 2, "Deployment failed"
    assert log_output.counter == 0, "No output to log"

    monkeypatch.setattr(deploy_helper, "get_router_connection", get_fail_router)

    assert not deployment.load_and_execute_config_changes(
        [], DeployMetadata("abc", "def", "/app", "/status", {}), "abc123"
    ), "Send config files fails"

    assert add_log.counter == 2, "Single log added"
    assert fail_send_config_files.counter == 1, "Attempted to send files"
    assert fail_deployment.counter == 3, "Deployment failed"
    assert log_output.counter == 0, "No output to log"

    @counter_wrapper
    def sent_config_files(*args, **kwargs):
        """
        .
        """
        return "/commands.sh"

    monkeypatch.setattr(deployment, "send_config_files_to_router", sent_config_files)

    assert not deployment.load_and_execute_config_changes(
        [], DeployMetadata("abc", "def", "/app", "/status", {}), "abc123"
    ), "Exec command fails"

    assert add_log.counter == 4, "Two logs added"
    assert sent_config_files.counter == 1, "Attempted to send files"
    assert fail_deployment.counter == 4, "Deployment failed"
    assert log_output.counter == 0, "No output to log"

    monkeypatch.setattr(deploy_helper, "get_router_connection", get_success_router)

    assert deployment.load_and_execute_config_changes(
        [], DeployMetadata("abc", "def", "/app", "/status", {}), "abc123"
    ), "Exec command succeeds"

    assert add_log.counter == 6, "Two logs added"
    assert sent_config_files.counter == 2, "Config file sent"
    assert fail_deployment.counter == 4, "Deployment did not fail"
    assert log_output.counter == 1, "Output logged"


def test_handle_deployment(monkeypatch, capsys):
    """
    .
    """
    deploy_config = {
        "git": {
            "diff-config-folder": "/diff",
            "config-folder": "/config",
            "webhook-url": "/webhook",
        },
        "apply-difference-only": True,
    }

    form = {
        "deployment": {
            "statuses_url": "/statuses",
            "ref": "fed",
            "payload": {"previous_commit": "abc"},
        },
        "action": "pending",
        "repository": {"clone_url": "/clone"},
    }

    assert deployment.handle_deployment(
        form, deploy_config, "abc123"
    ), "Non created action successful"
    printed = capsys.readouterr()
    assert (
        printed.out == "Ignoring deployment action pending\n"
    ), "Ignore message prints"

    # pylint: disable=unused-argument,too-many-arguments
    @counter_wrapper
    def update_deploy_state(
        status_url: str,
        external_app_url: str,
        before: str,
        after: str,
        access_token: str,
        result: str,
    ):
        """
        .
        """
        state = (
            "in_progress"
            if update_deploy_state.counter % 2
            else ("success" if update_deploy_state.counter > 2 else "failure")
        )
        assert result == state, "Deploy state is set as expected"

    @counter_wrapper
    def setup_repos(*args, **kwargs):
        """
        .
        """

    @counter_wrapper
    def get_commands(*args, **kwargs):
        """
        .
        """
        return []

    @counter_wrapper
    def load_execute(*args, **kwargs):
        """
        .
        """
        return load_execute.counter == 2

    form["action"] = "created"

    monkeypatch.setattr(db, "update_deployment_status", lambda *args, **kwargs: False)
    monkeypatch.setattr(api, "update_deployment_state", update_deploy_state)
    monkeypatch.setattr(api, "setup_config_repo", setup_repos)
    monkeypatch.setattr(deploy_helper, "get_commands_to_run", get_commands)
    monkeypatch.setattr(deployment, "load_and_execute_config_changes", load_execute)
    assert not deployment.handle_deployment(
        form, deploy_config, "abc123"
    ), "Fail to update deployment"
    printed = capsys.readouterr()
    assert printed.out == (
        "Failed to update local copy of deploy to in progress\n"
        "Failed to update local copy of deploy to failure\n"
    ), "Failure message printed"

    assert update_deploy_state.counter == 2, "Deploy set to in progress and finalized"
    assert setup_repos.counter == 1, "Repos set up"
    assert get_commands.counter == 1, "Commands retrieved"
    assert load_execute.counter == 1, "Attempted to run commands"

    monkeypatch.setattr(db, "update_deployment_status", lambda *args, **kwargs: True)
    assert deployment.handle_deployment(
        form, deploy_config, "abc123"
    ), "Deployment succeeds"
    printed = capsys.readouterr()
    assert printed.out == "", "No messages printed"

    assert update_deploy_state.counter == 4, "Deploy set to in progress and finalized"
    assert setup_repos.counter == 2, "Repos set up"
    assert get_commands.counter == 2, "Commands retrieved"
    assert load_execute.counter == 2, "Attempted to run commands"
