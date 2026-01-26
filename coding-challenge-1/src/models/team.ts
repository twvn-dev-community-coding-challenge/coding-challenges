import Member, { MemberId } from "./member.js";

export default class Team {
    public readonly members: Member[];
    private lastSelectedId: MemberId | null;

    constructor(members: Member[], lastSelectedMemberId: MemberId | null = null) {
        this.members = [...members];
        this.lastSelectedId = lastSelectedMemberId;
    }

    get lastSelectedMemberId(): MemberId | null {
        return this.lastSelectedId;
    }

    setLastSelectedMemberId(memberId: MemberId | null): void {
        this.lastSelectedId = memberId;
    }

    getActiveMembers(): Member[] {
        return this.members.filter((member) => member.isActive);
    }

    getMemberById(memberId: MemberId): Member | undefined {
        return this.members.find((member) => member.id === memberId);
    }
}
