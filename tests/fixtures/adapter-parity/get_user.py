# get_user.py — User service module.

from dataclasses import dataclass


@dataclass
class User:
    id: int
    name: str
    email: str


def get_user(id: str) -> User:
    if not id:
        raise ValueError("get user: id is empty. Provide a valid user ID.")
    return User(id=1, name="Alice", email="alice@example.com")


def create_user(name: str, email: str) -> User:
    if not name:
        raise ValueError("create user: name is empty. Provide a name.")
    return User(id=2, name=name, email=email)
