import pytest

from member_management.member import Member
from member_management.status import Status
from member_management.member_manager import MemberManager

# import sys
# print(f"Sys path: {sys.path}")


@pytest.fixture
def four_active_members():
    return [
        Member("Alice", Status.isActive),
        Member("Bob", Status.isActive),
        Member("Charlie", Status.isActive),
        Member("Diana", Status.isActive),
    ]


@pytest.fixture(scope="module")
def member_manager_object():
    return MemberManager()
