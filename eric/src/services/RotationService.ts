import { Member, Result } from '../utils/types';
import { RotationStrategy } from '../strategies/RotationStrategy';
import { getMemberlist } from '../data/members';

/**
 * Service Layer: Core business logic for team member rotation
 * 
 * Implements Singleton Pattern to ensure:
 * - Single source of truth for rotation state across all requests
 * - Consistent tracking of last selected member
 * - No duplicate instances causing state conflicts
 * 
 * Responsibilities:
 * - Validate rotation parameters
 * - Delegate selection to rotation strategy
 * - Maintain rotation state (last selected member tracking)
 * - Return structured results with success/failure status
 */
export class RotationService {
    /** Singleton instance - ensures only one service exists application-wide */
    private static instance: RotationService;

    /** 
     * Cached index of the last selected member for O(1) lookup
     * Used to resume rotation from the correct position
     */
    private lastSelectedMemberIndex: number | null = null;

    /** 
     * ID of the last selected member for sync verification
     * Used to detect and recover from index/ID desynchronization
     * (e.g., when member list changes between requests)
     */
    private lastSelectedMemberId: number | null = null;

    /**
     * Gets or creates the singleton instance of RotationService
     * 
     * Uses lazy initialization - instance is created only on first call
     * All subsequent calls return the same instance
     * 
     * @returns The singleton RotationService instance
     */
    static getInstance(): RotationService {
        if (!RotationService.instance) {
            RotationService.instance = new RotationService();
        }
        return RotationService.instance;
    }

    /**
     * Selects the next member(s) for rotation using the provided strategy
     * 
     * Validation Steps:
     * 1. Ensure count is at least 1
     * 2. Ensure count doesn't exceed total members
     * 3. Ensure enough active members are available
     * 
     * @param strategy - The rotation strategy to use (e.g., RoundRobinStrategy)
     * @param count - Number of members to select (default: 1)
     * @returns Result object with success status and selected members or error message
     * 
     * @example
     * // Get next 2 members using round-robin
     * const result = service.getNext(roundRobinStrategy, 2);
     * if (result.success) {
     *   console.log(result.data); // [{ id: 1, name: 'Alice' }, { id: 2, name: 'Bob' }]
     * }
     */
    getNext(strategy: RotationStrategy, count: number = 1): Result<Omit<Member, 'isActive'>[]> {

        // Validation: count must be positive
        if (count < 1) {
            return {
                success: false,
                message: 'Count must be greater than 0',
            };
        }

        // Load current member list from data source
        const memberList = getMemberlist();

        // Validation: count cannot exceed total members (regardless of active status)
        if (count > memberList.length) {
            return {
                success: false,
                message: `Count cannot exceed total number of members (${memberList.length})`,
            };
        }

        // Delegate selection to the strategy (Strategy Pattern)
        // Pass current state for rotation continuity
        const selectedMembers = strategy.getNext(
            memberList,
            this.lastSelectedMemberIndex,
            this.lastSelectedMemberId,
            count
        );

        // Validation: check if requested count exceeds available active members
        if (count > selectedMembers?.members.length!) {
            return {
                success: false,
                message: `Not enough active members to fulfill the request. Requested: ${count}, Available: ${selectedMembers?.members.length}`,
            };
        }

        // Handle case where no active members exist
        if (selectedMembers === null) {
            return {
                success: false,
                message: 'No active members available',
            };
        }

        // Update rotation state for next request
        // Storing both index and ID enables:
        // - O(1) lookup via index for normal operations
        // - ID-based recovery when member list changes (self-healing)
        this.lastSelectedMemberIndex = selectedMembers.lastMemberIndex;
        this.lastSelectedMemberId = selectedMembers.lastMemberId;

        return {
            success: true,
            data: selectedMembers.members
        };
    }

}
