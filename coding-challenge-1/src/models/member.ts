export type MemberId = string | number;

type MemberParams = {
    id: MemberId;
    name: string;
    isActive: boolean;
};

export default class Member {
    public readonly id: MemberId;
    public readonly name: string;
    private active: boolean;

    constructor({ id, name, isActive }: MemberParams) {
        this.id = id;
        this.name = name;
        this.active = isActive;
    }

    get isActive(): boolean {
        return this.active;
    }

    setActive(isActive: boolean): void {
        this.active = isActive;
    }
}
