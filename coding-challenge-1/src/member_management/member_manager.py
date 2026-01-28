from typing import Dict, List

from member_management.member import Member
from member_management.status import Status


class MemberManager:
    def __init__(self, members: Dict[int, Member] = {}):
        self.members = members
        self.counter = 0
        self.member_count = 0
        self.active_count = 0
        self.prev_id = None

    
    def get_member(self, member_id: int) -> Member:
        if member_id not in self.members:
            raise KeyError("User not exist")
        return self.members[member_id]
    

    def add_member(self, member_info: Member) -> None:
        print("Current members: ", self.members)
        if member_info.id in self.members:
            raise KeyError("Member already exists")   
        else:
            member_info.id = self.member_count + 1
            self.members[member_info.id] = member_info
            self.member_count += 1
            self.active_count += 1


    def change_status(self, member_id: int, status: Status) -> None:
        if member_id not in self.members:
            raise KeyError("User not exist")
        
        member = self.members[member_id]
        if member.status == status:
            # Same status
            return

        # Assign status
        member.status = status
        
        if status == Status.isActive:
            self.active_count += 1
            return
        
        if status == Status.isNotActive:
            self.active_count -= 1
            return

    
    def get_next(self, n: int = 1) -> List[Member]:
        """
        Several considerations:
            n > active members: loop min(n, active members)
            n < active members: loop min(n, active members)
        """
        # No active member to return
        # if self.active_count == 0:
        #     return []
        
        # if self.active_count == 1:
        #     return [self.members[self.counter % self.member_count + 1]]
        
        i = min(n, self.active_count)
        ans = []

        while i:
            member_id = self.counter % self.member_count + 1
            # print(f"Getting member id: {member_id}")
            member = self.members[member_id]

            if member.status == Status.isActive:
                ans.append(member)
                self.prev_id = member_id
                i -= 1
            self.counter += 1

        return ans
