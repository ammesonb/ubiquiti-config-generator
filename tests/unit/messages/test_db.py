"""
Test the DB functionality
"""
import os
from os import path
import sqlite3

from ubiquiti_config_generator.messages import db
from ubiquiti_config_generator.messages.check import Check
from ubiquiti_config_generator.messages.deployment import Deployment
from ubiquiti_config_generator.messages.log import Log
from ubiquiti_config_generator.testing_utils import counter_wrapper


def get_test_db_file() -> str:
    """
    Generates a random test db with the PID in it
    """
    return "messages_test." + os.urandom(10).hex() + str(os.getpid()) + ".db"


def remove_db_file(file_path: str):
    """
    Removes the db file
    """
    if path.isfile(file_path):
        os.remove(file_path)

    assert not path.isfile(file_path), "Test DB removed"


def test_initialize():
    """
    .
    """
    test_db = get_test_db_file()

    try:
        cursor = db.initialize_db(test_db)
        cursor.execute(
            """
            SELECT name
            FROM   sqlite_master
            WHERE  type = 'table'
            AND    name IN (
                'commit_check',
                'check_log',
                'deployment',
                'deployment_log'
            )
            """
        )
        results = cursor.fetchall()
        assert len(results) == 4, "All tables initialized"
    finally:
        remove_db_file(test_db)


def test_get_cursor(monkeypatch):
    """
    .
    """
    # pylint: disable=too-few-public-methods,unused-argument
    class FakeConnection:
        """
        A connection
        """

        row_factory = None

        @counter_wrapper
        def cursor(self):
            """
            Would open a cursor
            """
            return self

    @counter_wrapper
    def connect(db_file: str):
        """
        .
        """
        return FakeConnection()

    @counter_wrapper
    def initialize(db_file: str):
        """
        .
        """
        return connect(db_file).cursor()

    monkeypatch.setattr(sqlite3, "connect", connect)
    monkeypatch.setattr(path, "isfile", lambda db_path: True)
    monkeypatch.setattr(db, "initialize_db", initialize)

    cursor = db.get_cursor("a file")
    assert cursor.row_factory == sqlite3.Row, "Row factory set"
    assert FakeConnection.cursor.counter == 1, "Cursor called"
    assert initialize.counter == 0, "DB not initialized if it exists"

    monkeypatch.setattr(path, "isfile", lambda db_path: False)
    cursor = db.get_cursor("another file")
    assert cursor.row_factory == sqlite3.Row, "Row factory set"
    assert FakeConnection.cursor.counter == 2, "Cursor called again"
    assert initialize.counter == 1, "Initialize called"


def test_check(monkeypatch):
    """
    .
    """
    check1 = Check("123abc", "pending", 1611608732)
    check2 = Check("abc123", "pending", 1611608732, 1611608832, ["stuff", "things"])

    monkeypatch.setattr(db, "get_check_logs", lambda revision, cursor=None: [])

    db_file = get_test_db_file()
    try:
        cursor = db.initialize_db(db_file)

        # pylint: disable=unused-argument
        @counter_wrapper
        def get_cursor(db_file: str = ""):
            """
            .
            """
            return cursor

        monkeypatch.setattr(db, "get_cursor", get_cursor)

        assert db.create_check(check1, cursor), "First check added successfully"
        assert get_cursor.counter == 0, "Cursor not retrieved if passed in"
        assert db.create_check(check2), "Second check added"
        assert get_cursor.counter == 1, "Cursor retrieved if not passed in"

        assert db.get_check("123abc") == check1, "First check retrieved"
        monkeypatch.setattr(
            db, "get_check_logs", lambda revision, cursor=None: ["stuff", "things"]
        )
        assert db.get_check("abc123") == check2, "Second check retrieved"
    finally:
        remove_db_file(db_file)


def test_check_logs(monkeypatch):
    """
    .
    """
    log1 = Log(
        revision1="abc123", message="foo", status="log", utc_unix_timestamp=1611608732
    )
    log2 = Log(
        revision1="abc123",
        message="bar",
        status="success",
        utc_unix_timestamp=1611608732,
    )

    db_file = get_test_db_file()
    try:
        cursor = db.initialize_db(db_file)

        # pylint: disable=unused-argument
        @counter_wrapper
        def get_cursor(db_file: str = ""):
            """
            .
            """
            return cursor

        monkeypatch.setattr(db, "get_cursor", get_cursor)

        assert db.add_check_log(log1, cursor), "First check added successfully"
        assert get_cursor.counter == 0, "Cursor not retrieved if passed in"
        assert db.add_check_log(log2), "Second check added"
        assert get_cursor.counter == 1, "Cursor retrieved if not passed in"

        logs = db.get_check_logs("abc123", cursor)
        assert get_cursor.counter == 1, "Cursor not retrieved if passed in"
        assert len(logs) == 2, "Both logs retrieved"
        assert logs[0] == log1, "First log retrieved"
        assert logs[1] == log2, "Second log retrieved"
    finally:
        remove_db_file(db_file)


def test_deployment(monkeypatch):
    """
    .
    """
    deploy1 = Deployment("987abc", "876abc", "pending", 1611708732)
    deploy2 = Deployment(
        "abc987", "abc876", "pending", 1611708732, 1611708832, ["stuff", "things"]
    )

    monkeypatch.setattr(
        db, "get_deployment_logs", lambda revision1, revision2, cursor=None: []
    )

    db_file = get_test_db_file()
    try:
        cursor = db.initialize_db(db_file)

        # pylint: disable=unused-argument
        @counter_wrapper
        def get_cursor(db_file: str = ""):
            """
            .
            """
            return cursor

        monkeypatch.setattr(db, "get_cursor", get_cursor)

        assert db.create_deployment(
            deploy1, cursor
        ), "First deployment added successfully"
        assert get_cursor.counter == 0, "Cursor not retrieved if passed in"
        assert db.create_deployment(deploy2), "Second deployment added"
        assert get_cursor.counter == 1, "Cursor retrieved if not passed in"

        assert (
            db.get_deployment("987abc", "876abc") == deploy1
        ), "First deployment retrieved"
        monkeypatch.setattr(
            db,
            "get_deployment_logs",
            lambda revision1, revision2, cursor=None: ["stuff", "things"],
        )
        assert (
            db.get_deployment("abc987", "abc876") == deploy2
        ), "Second deployment retrieved"
    finally:
        remove_db_file(db_file)


def test_deployment_logs(monkeypatch):
    """
    .
    """
    log1 = Log(
        revision1="bcd123",
        revision2="cde234",
        message="ipsum",
        status="log",
        utc_unix_timestamp=1611608732,
    )
    log2 = Log(
        revision1="bcd123",
        revision2="cde234",
        message="lorem",
        status="failure",
        utc_unix_timestamp=1611608732,
    )

    db_file = get_test_db_file()
    try:
        cursor = db.initialize_db(db_file)

        # pylint: disable=unused-argument
        @counter_wrapper
        def get_cursor(db_file: str = ""):
            """
            .
            """
            return cursor

        monkeypatch.setattr(db, "get_cursor", get_cursor)

        assert db.add_deployment_log(
            log1, cursor
        ), "First deployment added successfully"
        assert get_cursor.counter == 0, "Cursor not retrieved if passed in"
        assert db.add_deployment_log(log2), "Second deployment added"
        assert get_cursor.counter == 1, "Cursor retrieved if not passed in"

        logs = db.get_deployment_logs("bcd123", "cde234", cursor)
        assert get_cursor.counter == 1, "Cursor not retrieved if passed in"
        assert len(logs) == 2, "Both logs retrieved"
        assert logs[0] == log1, "First log retrieved"
        assert logs[1] == log2, "Second log retrieved"
    finally:
        remove_db_file(db_file)


def test_missing_entries(monkeypatch):
    """
    .
    """
    db_file = get_test_db_file()
    try:
        cursor = db.initialize_db(db_file)
        assert db.get_check("missing", cursor) == Check(
            "missing", "nonexistent", 0, 0, None
        ), "Missing check revision is correct"

        missing_check_logs = db.get_check_logs("missing", cursor)
        assert isinstance(missing_check_logs, list) and not len(
            missing_check_logs
        ), "Missing check logs is empty list"

        assert db.get_deployment("missing", "also missing", cursor) == Deployment(
            "missing", "also missing", "nonexistent", 0, 0, None
        ), "Missing deployment revision is correct"

        missing_deployment_logs = db.get_deployment_logs(
            "missing", "also missing", cursor
        )
        assert isinstance(missing_deployment_logs, list) and not len(
            missing_deployment_logs
        ), "Missing deployment logs is empty list"

        assert not db.update_check_status(
            Log("missing", "a message", status="success"), cursor
        ), "Can't update a check that does not exist"
        assert not db.get_check_logs(
            "missing", cursor
        ), "No logs added for nonexistent check"

        assert not db.update_deployment_status(
            Log("missing", "a message", revision2="also missing", status="success"),
            cursor,
        ), "Can't update a deployment that does not exist"
        assert not db.get_deployment_logs(
            "missing", "also missing", cursor
        ), "No logs added for nonexistent deployment"
    finally:
        remove_db_file(db_file)


def test_update_check(monkeypatch):
    """
    .
    """
    log = Log(
        revision1="abc123",
        message="bar",
        status="success",
        utc_unix_timestamp=1611608752.0,
    )
    check = Check("abc123", "pending", 1611608732.0, logs=[log])

    db_file = get_test_db_file()
    try:
        cursor = db.initialize_db(db_file)

        # pylint: disable=unused-argument
        @counter_wrapper
        def get_cursor(db_file: str = ""):
            """
            .
            """
            return cursor

        monkeypatch.setattr(db, "get_cursor", get_cursor)

        assert db.create_check(check), "First check added successfully"
        assert db.update_check_status(log), "Check updated"
        assert get_cursor.counter == 2, "Cursor retrieved if not passed in"

        check.status = "success"
        check.ended_at = 1611608752.0
        assert db.get_check_logs("abc123", cursor) == [log], "Log added"
        assert db.get_check("abc123") == check, "Check updated"
    finally:
        remove_db_file(db_file)


def test_update_deployment(monkeypatch):
    """
    .
    """
    log = Log(
        revision1="abc123",
        revision2="cba123",
        message="foo",
        status="success",
        utc_unix_timestamp=1611608777.0,
    )
    deployment = Deployment("abc123", "cba123", "pending", 1611608700.0, logs=[log])

    db_file = get_test_db_file()
    try:
        cursor = db.initialize_db(db_file)

        # pylint: disable=unused-argument
        @counter_wrapper
        def get_cursor(db_file: str = ""):
            """
            .
            """
            return cursor

        monkeypatch.setattr(db, "get_cursor", get_cursor)

        assert db.create_deployment(deployment), "First deployment added successfully"
        assert db.update_deployment_status(log), "Deployment updated"
        assert get_cursor.counter == 2, "Cursor retrieved if not passed in"

        deployment.status = "success"
        deployment.ended_at = 1611608777.0
        assert db.get_deployment_logs("abc123", "cba123", cursor) == [log], "Log added"
        assert db.get_deployment("abc123", "cba123") == deployment, "Deployment updated"
    finally:
        remove_db_file(db_file)
