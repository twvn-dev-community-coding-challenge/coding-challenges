import { createMember } from "./member.js";
import Team from "./team.js";

describe("Team", () => {
    it("returns only active members", () => {
        const members = [
            createMember(1, "Alice", true),
            createMember(2, "Bob", false),
            createMember(3, "Charlie", true),
        ];
        const team = new Team(members);

        const activeMembers = team.getActiveMembers();

        expect(activeMembers.map((member) => member.id)).toEqual([1, 3]);
    });

    it("tracks the last selected member id", () => {
        const members = [createMember(1, "Diana", true)];
        const team = new Team(members);

        team.setLastSelectedMemberId(1);

        expect(team.lastSelectedMemberId).toBe(1);
    });
});
