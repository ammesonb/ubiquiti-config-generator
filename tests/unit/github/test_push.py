"""
Test push Git code
"""

from ubiquiti_config_generator.github import push, api
from ubiquiti_config_generator.messages import db
from ubiquiti_config_generator.testing_utils import counter_wrapper
from tests.unit.github.test_api import Response


def test_is_against_primary_branch():
    """
    .
    """
    config = {"git": {"primary-branch": "main"}}
    assert not push.is_against_primary_branch(
        config, "refs/head/feature"
    ), "Feature is not primary"
    assert not push.is_against_primary_branch(
        config, "refs/head/main-test"
    ), "Test from main is not primary"
    assert push.is_against_primary_branch(config, "refs/head/main"), "Main is primary"


def test_check_for_deployment(monkeypatch, capsys):
    """
    .
    """
    # pylint: disable=unused-argument
    @counter_wrapper
    def set_commit_status(url: str, ref: str, access_token: str, status: str):
        """
        .
        """
        assert status == "pending", "Commit status is pending"

    @counter_wrapper
    def create_deployment(*args, **kwargs):
        """
        .
        """

    @counter_wrapper
    def update_deploy_status(log):
        """
        .
        """
        assert log.status == "failure", "Should update to failed"

    @counter_wrapper
    def add_deploy_log(log):
        """
        .
        """
        assert log.status == "success", "Deployment created log is success"
        assert log.message == "Deployment created", "Message is correct"

    monkeypatch.setattr(api, "set_commit_status", set_commit_status)
    monkeypatch.setattr(
        push, "is_against_primary_branch", lambda *args, **kwargs: False
    )
    monkeypatch.setattr(db, "create_deployment", create_deployment)
    monkeypatch.setattr(db, "update_deployment_status", update_deploy_status)
    monkeypatch.setattr(db, "add_deployment_log", add_deploy_log)

    form = {
        "repository": {"statuses_url": "/statuses", "deployments_url": "/deployments"},
        "before": "bef",
        "after": "af",
        "ref": "123abc",
    }

    assert (
        push.check_push_for_deployment({}, form, "abc123") is None
    ), "No deployments against non-primary branch"
    assert set_commit_status.counter == 1, "Commit status set, against all branches"

    monkeypatch.setattr(push, "is_against_primary_branch", lambda *args, **kwargs: True)

    @counter_wrapper
    def send_github_request(url: str, method: str, access_token: str, json: dict):
        """
        .
        """
        assert json["payload"]["previous_commit"] == (
            "sha" if send_github_request.counter == 1 else "bef"
        ), "Previous commit correct"
        return (
            Response({"message": "failed"}, 403)
            if send_github_request.counter == 1
            else Response({"message": "created"}, 201)
        )

    monkeypatch.setattr(api, "send_github_request", send_github_request)
    monkeypatch.setattr(api, "get_active_deployment_sha", lambda *args, **kwargs: "sha")

    assert not push.check_push_for_deployment(
        {}, form, "abc123"
    ), "Create deployment fails"
    printed = capsys.readouterr()
    assert set_commit_status.counter == 2, "Commit status set"
    assert send_github_request.counter == 1, "Attempted to create deployment"
    assert printed.out == (
        "Failed to create deployment\n" "{'message': 'failed'}\n"
    ), "Fail deploy printed"
    assert create_deployment.counter == 1, "Deployment created in db"
    assert update_deploy_status.counter == 1, "Attempted to update deployment status"

    monkeypatch.setattr(api, "get_active_deployment_sha", lambda *args, **kwargs: None)
    assert push.check_push_for_deployment(
        {}, form, "abc123"
    ), "Create deployment succeeds"
    printed = capsys.readouterr()
    assert set_commit_status.counter == 3, "Commit status set"
    assert send_github_request.counter == 2, "Deployment created"
    assert printed.out == "", "Nothing printed"
    assert add_deploy_log.counter == 1, "Added log for deployment"
