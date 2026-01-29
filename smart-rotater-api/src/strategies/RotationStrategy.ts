import { Member } from '../utils/types';

/**
 * Strategy Pattern: Interface for rotation algorithms
 * This allows different rotation strategies to be plugged in
 */
export interface RotationStrategy {
    /**
     * Get the next member(s) for rotation
     * @param members - List of team members
     * @param lastMemberId - ID of the last selected member (or null if none)
     * @param count - Number of members to return
     * @returns Array of selected members
     */
    getNext(members: Member[], lastMemberIndex: number | null, count: number): { lastIndex: number | null, members: Omit<Member, 'isActive'>[] } | null;
}

/**
 * Round-Robin Rotation Strategy
 * Implements fair rotation by cycling through active members in order
 */
export class RoundRobinStrategy implements RotationStrategy {
    private currentIndex: number = 0;

    getNext(members: Member[], lastMemberIndex: number | null, count: number): { lastIndex: number | null, members: Omit<Member, 'isActive'>[] } | null {
        const result: Omit<Member, 'isActive'>[] = [];
        const numberLength = members.length;
        let checkedNumber = 0;

        if (lastMemberIndex === null) {
            this.currentIndex = 0;
        }

        while (result.length < count && checkedNumber < numberLength) {
            if (members[this.currentIndex].isActive) {
                result.push({
                    id: members[this.currentIndex].id,
                    name: members[this.currentIndex].name,
                });
            }
            this.currentIndex = (this.currentIndex + 1) % numberLength;
            checkedNumber += 1;
        }

        if (result.length === 0) {
            return null
        }

        return {
            lastIndex: this.currentIndex - 1 < 0 ? numberLength - 1 : this.currentIndex - 1,
            members: result
        };
    }
}

export class RotationStrategyFactory {
    private static roundRobinStrategyInstance: RotationStrategy;

    static getRoundRobinStrategy(): RotationStrategy {
        if (!this.roundRobinStrategyInstance) {
            this.roundRobinStrategyInstance = new RoundRobinStrategy();
        }
        return this.roundRobinStrategyInstance;
    }
}
