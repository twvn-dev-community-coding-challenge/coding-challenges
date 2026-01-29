import { Member, Result } from '../utils/types';
import { RotationStrategy } from '../strategies/RotationStrategy';
import { getMemberlist } from '../data/members';

/**
 * Service layer for rotation logic
 * Handles business logic for member rotation
 * Implements Singleton Pattern to ensure only one instance exists
 */
export class RotationService {
    private static instance: RotationService;
    private lastSelectedMemberIndex: number | null = null;

    /**
     * Get or create the singleton instance
     * @param members - Team members (only used on first instantiation)
     * @param strategy - Rotation strategy (only used on first instantiation)
     */
    static getInstance(): RotationService {
        if (!RotationService.instance) {
            RotationService.instance = new RotationService();
        }
        return RotationService.instance;
    }


    /**
     * Get the next member(s) for rotation
     * @param count - Number of members to return
     * @returns RotationResult with selected members
     */
    getNext(strategy: RotationStrategy, count: number = 1): Result<Omit<Member, 'isActive'>[]> {

        if (count < 1) {
            return {
                success: false,
                message: 'Count must be greater than 0',
            };
        }

        const memberList = getMemberlist();

        if (count > memberList.length) {
            return {
                success: false,
                message: `Count cannot exceed total number of members (${memberList.length})`,
            };
        }

        const selectedMembers = strategy.getNext(
            memberList,
            this.lastSelectedMemberIndex,
            count
        );

        if(count > selectedMembers?.members.length!) {
            return {
                success: false,
                message: `Not enough active members to fulfill the request. Requested: ${count}, Available: ${selectedMembers?.members.length}`,
            };
        }

        if (selectedMembers === null) {
            return {
                success: false,
                message: 'No active members available',
            };
        }


        this.lastSelectedMemberIndex = selectedMembers.lastIndex;


        return {
            success: true,
            data: selectedMembers.members
        };
    }

}
