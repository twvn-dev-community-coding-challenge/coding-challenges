import pytest

from fixtures import member_manager_object

from member_management.member_manager import MemberManager
from member_management.member import Member
from member_management.status import Status

# @pytest.mark.skip
def test_changes_status(member_manager_object: MemberManager):
    member_manager_object.add_member(Member("Alice", Status.isActive))

    member_manager_object.change_status(1, Status.isNotActive)
    assert member_manager_object.get_member(1).status == Status.isNotActive

    assert member_manager_object.member_count == 1
    assert member_manager_object.active_count == 0

    member_manager_object.change_status(1, Status.isActive)
    assert member_manager_object.get_member(1).status == Status.isActive

    member_manager_object.change_status(1, Status.isActive)
    assert member_manager_object.get_member(1).status == Status.isActive

    assert member_manager_object.active_count == 1


    # Asset non-existence member
    with pytest.raises(KeyError):
        member_manager_object.change_status(2, Status.isActive)
