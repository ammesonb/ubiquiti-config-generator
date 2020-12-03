"""
Handles setup for the module
"""
import setuptools

with open("README.md", "r") as fh:
    long_description = fh.read()

setuptools.setup(
    name="ubiquiti-config-generator",
    version="1.0.0",
    author="Brett Ammeson",
    author_email="ammesonb@gmail.com",
    description=("Dynamically generates and applies ubiquiti configurations"),
    long_description=long_description,
    url="https://github.com/ammesonb/ubiquiti-config-generator",
    packages=setuptools.find_packages(),
    python_requires=">=3.7",
)
