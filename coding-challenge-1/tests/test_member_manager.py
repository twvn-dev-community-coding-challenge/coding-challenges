import pytest

from fixtures import (
    four_active_members,
    member_manager_object    
)

from member_management.member_manager import MemberManager
from member_management.member import Member
from member_management.status import Status

# @pytest.mark.skip
def test_returns_members_in_rotation_order(four_active_members: list[Member], member_manager_object: MemberManager):
    member_manager_object = MemberManager()

    for member in four_active_members:
        member_manager_object.add_member(member)

    results = member_manager_object.get_next(6)
    assert member_manager_object.active_count == 4

    assert len(results) == 4
    assert [result.name for result in results] == [
        "Alice",
        "Bob",
        "Charlie",
        "Diana",
    ]

    results = member_manager_object.get_next(3)
    assert [result.name for result in results] == [
        "Alice",
        "Bob",
        "Charlie"
    ]


    results = member_manager_object.get_next(2)
    assert [result.name for result in results] == [
        "Diana",
        "Alice"
    ]

    results = member_manager_object.get_next(3)
    assert [result.name for result in results] == [
        "Bob",
        "Charlie",
        "Diana"
    ]

    results = member_manager_object.get_next()
    assert [result.name for result in results] == [
        "Alice",
    ]



def test_add_member_with_dup_id(member_manager_object: MemberManager):
    member_manager_object.add_member(Member(1, "Alice", Status.isActive))

    with pytest.raises(KeyError):
        member_manager_object.add_member(Member(id=1, name="Bob", status=Status.isActive))