from setuptools import setup, find_packages

setup(
    name="DbLocust",
    packages=find_packages(where="locust_files"),
    package_dir={"": "locust_files"},
)
