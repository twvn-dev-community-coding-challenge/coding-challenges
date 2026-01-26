import Member from "./member.js";
import Team from "./team.js";

describe("Team", () => {
    it("returns only active members", () => {
        const members = [
            new Member({ id: 1, name: "Alice", isActive: true }),
            new Member({ id: 2, name: "Bob", isActive: false }),
            new Member({ id: 3, name: "Charlie", isActive: true }),
        ];
        const team = new Team(members);

        const activeMembers = team.getActiveMembers();

        expect(activeMembers.map((member) => member.id)).toEqual([1, 3]);
    });

    it("tracks the last selected member id", () => {
        const members = [new Member({ id: "a1", name: "Diana", isActive: true })];
        const team = new Team(members);

        team.setLastSelectedMemberId("a1");

        expect(team.lastSelectedMemberId).toBe("a1");
    });
});
