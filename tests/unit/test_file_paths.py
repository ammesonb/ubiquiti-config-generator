"""
Test the file paths loader functionality
"""
from os import path

from ubiquiti_config_generator import file_paths


def test_load_yaml_file():
    """
    Test the yaml file loader
    """
    result_obj = file_paths.load_yaml_from_file("tests/unit/test_yaml/test.yaml")
    assert result_obj == {
        "value": 1,
        "a-list": ["apple", "orange"],
        "nested": {"foo": "bar", "baz": ["ipsum", "lorem"]},
    }, "Yaml loaded correctly"


def test_get_config_files():
    """
    Check getting yaml files from a directory
    """
    paths = file_paths.get_config_files("tests/unit/test_yaml")
    assert len(paths) == 2, "Two files found"
    paths.sort()

    assert paths == [
        "tests/unit/test_yaml/example.yaml",
        "tests/unit/test_yaml/test.yaml",
    ], "Yaml files detected"


def test_get_config_directories():
    """
    Skip directories to find configuration yaml files
    """
    paths = file_paths.get_folders_with_config("tests/unit/test_yaml")
    assert len(paths) == 2, "Two paths found"
    paths.sort()

    assert paths == [
        "tests/unit/test_yaml/dir1/config.yaml",
        "tests/unit/test_yaml/dir2/config.yaml",
    ], "Paths detected"


def test_get_path():
    """
    Check the path joining
    """
    assert file_paths.get_path(["foo", "bar"]) == path.join(
        path.abspath("."), file_paths.TOP_LEVEL_DIRECTORY, "foo", "bar"
    ), "Path resolved correctly"
