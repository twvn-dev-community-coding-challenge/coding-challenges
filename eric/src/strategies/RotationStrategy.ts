import { Member } from '../utils/types';

/**
 * Strategy Pattern Interface: Defines the contract for rotation algorithms
 * 
 * This interface allows different rotation strategies (Round-Robin, Random, 
 * Weighted, etc.) to be implemented and swapped without changing the service layer
 */
export interface RotationStrategy {
    /**
     * Selects the next member(s) for rotation based on the algorithm
     * 
     * @param members - Complete list of team members (including inactive)
     * @param lastMemberIndex - Index of the last selected member in the array (null if first rotation)
     * @param lastMemberId - ID of the last selected member for sync verification (null if first rotation)
     * @param count - Number of members to select
     * @returns Object containing selected members and updated tracking info, or null if no active members
     */
    getNext(members: Member[], lastMemberIndex: number | null, lastMemberId: number | null, count: number): { lastMemberIndex: number | null, lastMemberId: number | null, members: Omit<Member, 'isActive'>[] } | null;
}

/**
 * Round-Robin Rotation Strategy
 * 
 * Implements fair rotation by cycling through active members sequentially.
 * This ensures each active member gets an equal number of turns over time.
 * 
 * Features:
 * - Skips inactive members automatically
 * - Handles index/ID desync for data integrity
 * - Wraps around to the beginning after reaching the end
 */
export class RoundRobinStrategy implements RotationStrategy {

    /**
     * Selects the next N active members in round-robin order
     * 
     * Algorithm:
     * 1. Determine starting index based on last selection
     * 2. Handle index/ID desync if detected (self-healing)
     * 3. Iterate through members, collecting active ones until count is reached
     * 4. Return selected members with updated tracking info
     * 
     * @param members - Complete list of team members
     * @param lastMemberIndex - Cached index of last selected member (for O(1) lookup)
     * @param lastMemberId - ID of last selected member (for sync verification)
     * @param count - Number of active members to return
     * @returns Selected members with tracking info, or null if no active members found
     * @throws Error if lastMemberId is set but not found in members list
     */
    getNext(members: Member[], lastMemberIndex: number | null, lastMemberId: number | null, count: number): { lastMemberIndex: number | null, lastMemberId: number | null, members: Omit<Member, 'isActive'>[] } | null {
        const result: Omit<Member, 'isActive'>[] = [];
        const totalMembers = members.length;
        let checkedCount = 0;
        let currentIndex;

        // Determine starting index for rotation
        if (lastMemberIndex === null) {
            // First rotation: start from beginning
            currentIndex = 0;
        } else if (members[lastMemberIndex].id !== lastMemberId) {
            // Index/ID desync detected: member list may have changed
            // Find the correct position by ID to maintain rotation integrity
            const correctLastIndex = members.findIndex(m => m.id === lastMemberId);
            if (correctLastIndex !== -1) {
                // Resume from the position after the last selected member
                currentIndex = (correctLastIndex + 1) % totalMembers;
            } else {
                // Last selected member no longer exists in the list
                throw new Error('not found lastMemberId in members list');
            }
        } else {
            // Normal case: continue from next position after last selection
            currentIndex = (lastMemberIndex + 1) % totalMembers;
        }

        // Collect active members until we have enough or checked all
        while (result.length < count && checkedCount < totalMembers) {
            if (members[currentIndex].isActive) {
                // Only include id and name in result (exclude isActive for cleaner response)
                result.push({
                    id: members[currentIndex].id,
                    name: members[currentIndex].name,
                });
            }
            // Move to next member (wrap around using modulo)
            currentIndex = (currentIndex + 1) % totalMembers;
            checkedCount += 1;
        }

        // No active members found after checking all
        if (result.length === 0) {
            return null
        }

        // Calculate the index of the last member we added to result
        // currentIndex has already moved past it, so we need to go back one position
        const lastIndex = currentIndex - 1 < 0 ? totalMembers - 1 : currentIndex - 1;

        return {
            lastMemberIndex: lastIndex,
            lastMemberId: members[lastIndex].id,
            members: result
        };
    }
}

/**
 * Factory Pattern: Centralizes strategy instance creation
 * 
 * Benefits:
 * - Ensures single instance of each strategy type (memory efficient)
 * - Provides a central point to add new strategy types
 * - Decouples strategy instantiation from usage
 */
export class RotationStrategyFactory {
    /** Cached instance of RoundRobinStrategy (Singleton per strategy type) */
    private static roundRobinStrategyInstance: RotationStrategy;

    /**
     * Returns the singleton instance of RoundRobinStrategy
     * Creates the instance on first call (lazy initialization)
     * 
     * @returns The RoundRobinStrategy singleton instance
     */
    static getRoundRobinStrategy(): RotationStrategy {
        if (!this.roundRobinStrategyInstance) {
            this.roundRobinStrategyInstance = new RoundRobinStrategy();
        }
        return this.roundRobinStrategyInstance;
    }
}
