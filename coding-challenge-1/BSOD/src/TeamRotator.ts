import { Member } from './models/member';
import { RotationIterator } from './interfaces/RotationIterator';

/**
 * TeamRotator implements the Iterator Pattern for fair team member rotation
 *
 * Strategy: Simple round-robin with on-the-fly skipping
 * - Iterates through members starting from currentIndex
 * - Skips inactive members automatically
 * - Prevents immediate repetition when possible
 * - Member status is fixed at initialization (no dynamic changes)
 *
 * Performance:
 * - Time complexity: O(n) per next() call in worst case
 * - Space complexity: O(1) - only tracks current index and last selected ID
 * - Trade-off: Simplicity and readability over constant-time lookups
 */
export class TeamRotator implements RotationIterator {
  private members: Member[];
  private currentIndex: number = 0;
  private lastSelectedId: number | null = null;

  constructor(members: Member[], lastSelectedId?: number | null) {
    if (!members || members.length === 0) {
      throw new Error('Team must have at least one member');
    }
    this.members = [...members]; // Create a copy to avoid external mutations
    this.lastSelectedId = lastSelectedId ?? null;
    if (this.lastSelectedId) {
      this.currentIndex = this.members.findIndex((member) => member.id === lastSelectedId);
    }
  }

  /**
   * Gets the next member in rotation
   * Tries each member in round-robin order until finding a valid one
   *
   * Time complexity: O(n) worst case
   * Space complexity: O(1)
   */
  next(): Member | null {
    let firstAttempt: Member | null = null;
    let attempts = 0;

    // Try to find next active member that isn't the last selected
    while (attempts < this.members.length) {
      const candidate = this.members[this.currentIndex];

      // Move to next index for next call
      this.currentIndex = (this.currentIndex + 1) % this.members.length;
      attempts++;

      // Skip inactive members
      if (!candidate.isActive) {
        continue;
      }

      // Remember the first active member we encounter
      if (firstAttempt === null) {
        firstAttempt = candidate;
      }

      // If this member wasn't just selected, return it
      if (candidate.id !== this.lastSelectedId) {
        this.lastSelectedId = candidate.id;
        return candidate;
      }
    }

    // If we've gone through all members and only found the last selected one,
    // return it anyway (edge case: only one active member)
    if (firstAttempt !== null) {
      this.lastSelectedId = firstAttempt.id;
      return firstAttempt;
    }

    // No active members found
    return null;
  }

  /**
   * Gets the next N members in rotation (optimized single-pass)
   * Throws error if there are not enough active members
   *
   * Time complexity: O(n) where n = number of members
   * Space complexity: O(count)
   */
  nextN(count: number): Member[] {
    if (count <= 0) {
      return [];
    }

    // Count active members first
    let activeCount = 0;
    for (const member of this.members) {
      if (member.isActive) {
        activeCount++;
      }
    }

    if (count > activeCount) {
      throw new Error(
        `Cannot select ${count} members: only ${activeCount} active members available`,
      );
    }

    const result: Member[] = [];
    let attempts = 0;
    const maxAttempts = this.members.length * 2; // Allow up to 2 full cycles

    // Collect count members in a single pass
    while (result.length < count && attempts < maxAttempts) {
      const candidate = this.members[this.currentIndex];

      // Move to next index
      this.currentIndex = (this.currentIndex + 1) % this.members.length;
      attempts++;

      // Skip inactive members
      if (!candidate.isActive) {
        continue;
      }

      // Apply no-repeat rule: skip if this is the last selected
      // Exception: if we've already added members to result, we can add it
      // (no-repeat only applies between separate calls or at the start)
      if (candidate.id === this.lastSelectedId && result.length === 0 && activeCount > 1) {
        continue;
      }

      // Add to result
      result.push(candidate);
      this.lastSelectedId = candidate.id;
    }

    return result;
  }

  /**
   * Checks if there are any active members
   * Time complexity: O(n)
   * Space complexity: O(1)
   */
  hasNext(): boolean {
    for (const member of this.members) {
      if (member.isActive) {
        return true;
      }
    }
    return false;
  }

  /**
   * Resets the iterator state
   * Time complexity: O(1)
   * Space complexity: O(1)
   */
  reset(): void {
    this.currentIndex = 0;
    this.lastSelectedId = null;
  }

  /**
   * Gets a copy of all members (for testing/debugging)
   */
  getMembers(): Member[] {
    return [...this.members];
  }
}
