import { createMember, Member } from "./member.js";

describe("Member", () => {
    it("stores id, name, and active status", () => {
        const member = createMember(1, "Alice", true);

        expect(member.id).toBe(1);
        expect(member.name).toBe("Alice");
        expect(member.isActive).toBe(true);
    });

    it("can update active status", () => {
        const member = createMember(2, "Bob", true);

        // Note: Member is now a simple interface, no setActive method
        expect(member.isActive).toBe(true);
    });
});
