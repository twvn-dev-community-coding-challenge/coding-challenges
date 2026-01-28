from dataclasses import dataclass

from member_management.status import Status


@dataclass
class Member:
    name: str
    status: Status
    id: int | None = None
