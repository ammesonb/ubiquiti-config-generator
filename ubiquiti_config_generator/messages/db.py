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
        CREATE TABLE IF NOT EXISTS commit_check (
            revision   VARCHAR(40) PRIMARY KEY,
            status     VARCHAR(20) NOT NULL,
            started_at FLOAT NOT NULL,
            ended_at   FLOAT NULL
        )
        """,
        """
        CREATE TABLE IF NOT EXISTS check_log (
            revision  VARCHAR(40) NOT NULL,
            status    VARCHAR(20) NOT NULL,
            timestamp FLOAT NOT NULL,
            message   TEXT NOT NULL,
            FOREIGN KEY (revision)
              REFERENCES commit_check (revision)
        )
        """,
        """
        CREATE TABLE IF NOT EXISTS deployment (
            from_revision VARCHAR(40),
            to_revision   VARCHAR(40),
            status        VARCHAR(20) NOT NULL,
            started_at    FLOAT NOT NULL,
            ended_at      FLOAT NULL,
            PRIMARY KEY (from_revision, to_revision),
            FOREIGN KEY (from_revision)
              REFERENCES commit_check (revision),
            FOREIGN KEY (to_revision)
              REFERENCES commit_check (revision)
        )
        """,
        """
        CREATE TABLE IF NOT EXISTS deployment_log (
            from_revision VARCHAR(40),
            to_revision   VARCHAR(40),
            status        VARCHAR(20) NOT NULL,
            timestamp     FLOAT NOT NULL,
            message       TEXT NOT NULL,
            FOREIGN KEY (from_revision)
              REFERENCES deployment (from_revision),
            FOREIGN KEY (to_revision)
              REFERENCES deployment (to_revision)
        )
        """,
    ]
    for statement in table_creation_statements:
        cursor.execute(statement)

    return get_cursor(db_file)


def get_cursor(db_file: str = DB_FILE) -> sqlite3.Cursor:
    """
    Opens a connection the db and gets a cursor
    """
    cursor = (
        initialize_db(db_file)
        if not path.isfile(db_file)
        else sqlite3.connect(db_file).cursor()
    )
    cursor.row_factory = sqlite3.Row
    return cursor


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

    return Check(**check, **{"logs": get_check_logs(revision, cursor)})


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
        **deployment,
        **{"logs": get_deployment_logs(from_revision, to_revision, cursor)}
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


def create_check(check: Check, cursor: Optional[sqlite3.Cursor] = None) -> bool:
    """
    Creates a check in the DB
    """
    cursor = cursor or get_cursor()
    result = cursor.execute(
        """
        INSERT INTO commit_check (
            revision,
            status,
            started_at,
            ended_at
        )
        VALUES (
            ?, ?, ?, ?
        )
        """,
        (check.revision, check.status, check.started_at, check.ended_at,),
    )

    return bool(result.lastrowid)


def update_check_status(log: Log, cursor: Optional[sqlite3.Cursor] = None) -> bool:
    """
    Updates the status of a check, adding a log with a reason
    """
    cursor = cursor or get_cursor()
    ended_at = log.utc_unix_timestamp if log.status in ["success", "failure"] else None
    result = cursor.execute(
        """
        UPDATE commit_check
        SET    status = ?,
               ended_at = ?
        WHERE  revision = ?
        """,
        (log.status, ended_at, log.revision1),
    )

    return result.rowcount and add_check_log(log, cursor)


def add_check_log(log: Log, cursor: Optional[sqlite3.Cursor] = None) -> bool:
    """
    Adds a log for a check message
    """
    cursor = cursor or get_cursor()
    result = cursor.execute(
        """
        INSERT INTO check_log (
            revision,
            status,
            timestamp,
            message
        ) VALUES (
            ?, ?, ?, ?
        )
        """,
        (log.revision1, log.status, log.utc_unix_timestamp, log.message,),
    )

    return bool(result.lastrowid)


def create_deployment(
    deployment: Deployment, cursor: Optional[sqlite3.Cursor] = None
) -> bool:
    """
    Creates a deployment in the DB
    """
    cursor = cursor or get_cursor()
    result = cursor.execute(
        """
        INSERT INTO deployment (
            from_revision,
            to_revision,
            status,
            started_at,
            ended_at
        )
        VALUES (
            ?, ?, ?, ?, ?
        )
        """,
        (
            deployment.from_revision,
            deployment.to_revision,
            deployment.status,
            deployment.started_at,
            deployment.ended_at,
        ),
    )

    return bool(result.lastrowid)


def update_deployment_status(log: Log, cursor: Optional[sqlite3.Cursor] = None) -> bool:
    """
    Updates the status of a deployment, adding a log with a reason
    """
    cursor = cursor or get_cursor()
    ended_at = log.utc_unix_timestamp if log.status in ["success", "failure"] else None
    result = cursor.execute(
        """
        UPDATE deployment
        SET    status = ?,
               ended_at = ?
        WHERE  from_revision = ?
        AND    to_revision = ?
        """,
        (log.status, ended_at, log.revision1, log.revision2,),
    )

    return result.rowcount and add_deployment_log(log, cursor)


def add_deployment_log(log: Log, cursor: Optional[sqlite3.Cursor] = None) -> bool:
    """
    Adds a log for a deployment message
    """
    cursor = cursor or get_cursor()
    result = cursor.execute(
        """
        INSERT INTO deployment_log (
            from_revision,
            to_revision,
            status,
            timestamp,
            message
        ) VALUES (
            ?, ?, ?, ?, ?
        )
        """,
        (
            log.revision1,
            log.revision2,
            log.status,
            log.utc_unix_timestamp,
            log.message,
        ),
    )

    return bool(result.lastrowid)
