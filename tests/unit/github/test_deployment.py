"""
Tests github deployment code
"""
import pytest

from ubiquiti_config_generator.github import api, deployment, deploy_helper
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
