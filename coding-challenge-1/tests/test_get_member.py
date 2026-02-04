import pytest

from fixtures import (
    member_manager_object    
)

from member_management.member_manager import MemberManager
from member_management.member import Member
from member_management.status import Status

# @pytest.mark.skip
def test_get_member(member_manager_object: MemberManager):
    member_manager_object.add_member(Member("Alice", Status.isActive))
    assert member_manager_object.get_member(1).name == "Alice"
    with pytest.raises(KeyError):
        member_manager_object.get_member(2)