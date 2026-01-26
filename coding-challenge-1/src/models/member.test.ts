import Member from "./member.js";

describe("Member", () => {
    it("stores id, name, and active status", () => {
        const member = new Member({ id: 1, name: "Alice", isActive: true });

        expect(member.id).toBe(1);
        expect(member.name).toBe("Alice");
        expect(member.isActive).toBe(true);
    });

    it("can update active status", () => {
        const member = new Member({ id: 2, name: "Bob", isActive: true });

        member.setActive(false);

        expect(member.isActive).toBe(false);
    });
});
