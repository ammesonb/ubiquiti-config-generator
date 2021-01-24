"""
Log database interaction functionality
"""
from os import path
import sqlite3
from typing import Optional, List

from ubiquiti_config_generator.messages.check import Check
from ubiquiti_config_generator.messages.deployment import Deployment
from ubiquiti_config_generator.messages.log import Log

DB_PATH = "messages"
DB_FILE = DB_PATH + "/messages.db"


def initialize_db(db_file: str = DB_FILE) -> sqlite3.Cursor:
    """
    Sets up the database for logging
    """
    connection = sqlite3.connect(db_file)
    cursor = connection.cursor()
    table_creation_statements = [
        """
        CREATE TABLE commit_check (
            revision   VARCHAR(40) PRIMARY KEY,
            status     VARCHAR(20) NOT NULL,
            started_at FLOAT NOT NULL,
            ended_at   FLOAT NULL
        )
        """,
        """
        CREATE TABLE check_log (
            revision  VARCHAR(40) NOT NULL,
            status    VARCHAR(20) NOT NULL,
            timestamp FLOAT NOT NULL,
            message   TEXT NOT NULL,
            FOREIGN KEY (revision)
              REFERENCES commit_check (revision)
        )
        """,
        """
        CREATE TABLE deployment (
            from_revision VARCHAR(40),
            to_revision   VARCHAR(40),
            status        VARCHAR(20) NOT NULL,
            started_at    FLOAT NOT NULL,
            ended_at      FLOAT NULL,
            PRIMARY KEY (from_revision, to_revision)
        )
        """,
        """
        CREATE TABLE deployment_log (
            from_revision VARCHAR(40),
            to_revision   VARCHAR(40),
            status        VARCHAR(20) NOT NULL,
            timestamp     FLOAT NOT NULL,
            message       TEXT NOT NULL,
            FOREIGN KEY (from_revision)
              REFERENCES commit_check (revision),
            FOREIGN KEY (to_revision)
              REFERENCES commit_check (revision)
        )
        """,
    ]
    for statement in table_creation_statements:
        cursor.execute(statement)
    return cursor


def get_cursor(db_file: str = DB_FILE) -> sqlite3.Cursor:
    """
    Opens a connection the db and gets a cursor
    """
    return (
        initialize_db(db_file)
        if not path.isfile(db_file)
        else sqlite3.connect(db_file).cursor()
    )


def get_check(revision: str, cursor: Optional[sqlite3.Cursor] = None) -> Check:
    """
    Get details about a revision check
    """
    cursor = cursor or get_cursor()
    result = cursor.execute(
        "SELECT * FROM commit_check WHERE revision = ?", (revision,)
    )
    check = result.fetchone()
    if not check:
        return Check(revision, "nonexistent", 0, 0, [])

    return Check(**check, logs=get_check_logs(revision, cursor))


def get_check_logs(revision: str, cursor: Optional[sqlite3.Cursor] = None) -> List[Log]:
    """
    Get the logs for a given check revision
    """
    cursor = cursor or get_cursor()
    result = cursor.execute("SELECT * FROM check_log WHERE revision = ?", (revision,))
    log_results = result.fetchall()
    return [
        Log(
            revision1=log["revision"],
            message=log["message"],
            status=log["status"],
            utc_unix_timestamp=log["timestamp"],
        )
        for log in log_results
    ]


def get_deployment(
    from_revision: str, to_revision: str, cursor: Optional[sqlite3.Cursor] = None
) -> Check:
    """
    Get details about a deployment
    """
    cursor = cursor or get_cursor()
    result = cursor.execute(
        """
        SELECT *
        FROM deployment
        WHERE from_revision = ?
        AND   to_revision = ?
        """,
        (from_revision, to_revision,),
    )
    deployment = result.fetchone()
    if not deployment:
        return Deployment(from_revision, to_revision, "nonexistent", 0, 0, [])

    return Deployment(
        **deployment, logs=get_deployment_logs(from_revision, to_revision, cursor)
    )


def get_deployment_logs(
    from_revision: str, to_revision: str, cursor: Optional[sqlite3.Cursor] = None
) -> List[Log]:
    """
    Get the logs for a given check revision
    """
    cursor = cursor or get_cursor()
    result = cursor.execute(
        """
        SELECT *
        FROM deployment_log
        WHERE from_revision = ?
        AND   to_revision = ?
        """,
        (from_revision, to_revision,),
    )
    log_results = result.fetchall()
    return [
        Log(
            revision1=log["from_revision"],
            revision2=log["to_revision"],
            message=log["message"],
            status=log["status"],
            utc_unix_timestamp=log["timestamp"],
        )
        for log in log_results
    ]
