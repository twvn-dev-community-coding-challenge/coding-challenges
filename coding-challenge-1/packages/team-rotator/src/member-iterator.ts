import { DuplicatedMemberIdentifierError, Member, NoActiveMembersError } from './types';

/**
 * Iterator Pattern Implementation
 *
 * Cycles through active members in a round-robin fashion.
 * This iterator ensures fair rotation by maintaining position
 * and automatically skipping inactive members.
 */
export class MemberIterator {
  private members: Member[];
  private currentIndex: number = 0;
  private hasStarted: boolean = false;

  constructor(members: Member[]) {
    this.validateMemberList(members);
    this.members = members;
  }

  /**
   * Add new member to the current rotator list (new member will come last)
   * @param member
   * @throws DuplicatedMemberIdentifierError if duplicated member id found from the list
   */
  addMember(member: Member): void {
    this.validateMemberList([...this.members, member]);
    this.members.push(member);
  }

  /**
   * Gets the next active member, skipping inactive ones
   * @param excludeId - Optional member ID to exclude (for no immediate repetition)
   * @returns The next active member, or null if none available
   */
  next(excludeId?: number): Member | null {
    const activeMembers = this.getActiveMembers();

    if (activeMembers.length === 0) {
      return null;
    }

    // If only one active member and it's the excluded one, return it anyway
    // (edge case: only one active member)
    if (activeMembers.length === 1) {
      return activeMembers[0];
    }

    // Find the starting position in the active members array
    let startIndex = this.findCurrentPositionInActive(activeMembers);

    // If we've already started, advance to next position for round-robin
    if (this.hasStarted) {
      startIndex = (startIndex + 1) % activeMembers.length;
    }

    // Start searching from the calculated position
    let attempts = 0;
    const maxAttempts = activeMembers.length;
    let searchIndex = startIndex;

    while (attempts < maxAttempts) {
      const member = activeMembers[searchIndex];

      // Skip if this is the excluded member (no immediate repetition)
      if (excludeId !== undefined && member.id === excludeId) {
        searchIndex = (searchIndex + 1) % activeMembers.length;
        attempts++;
        continue;
      }

      // Update current index to point to this member in the original array
      this.currentIndex = this.findMemberIndexInOriginal(member.id);
      this.hasStarted = true;
      return member;
    }

    // If all members are excluded (shouldn't happen with >1 active), return first
    return activeMembers[0];
  }

  /**
   * Gets the next N active members
   * @param count - Number of members to return
   * @param excludeId - Optional member ID to exclude
   * @returns Array of next N members
   */
  nextN(count: number, excludeId?: number): Member[] {
    const result: Member[] = [];
    const activeMembers = this.getActiveMembers();

    if (activeMembers.length === 0) {
      return result;
    }

    // If requesting more members than available, we'll cycle
    for (let i = 0; i < count; i++) {
      const member = this.next(
        i === 0 ? excludeId : undefined // Only exclude on first call
      );

      if (member) {
        result.push(member);
      } else {
        break; // No more members available
      }
    }

    return result;
  }

  private validateMemberList(members: Member[]) {
    if (members.length === 0 || members.every(({ isActive }) => !isActive)) {
      throw new NoActiveMembersError();
    }

    const memberIdsSet = new Set(members.map(({ id }) => id));
    if (members.length !== memberIdsSet.size) {
      throw new DuplicatedMemberIdentifierError();
    }
  }

  /**
   * Gets all active members
   */
  private getActiveMembers(): Member[] {
    return this.members.filter((m) => m.isActive);
  }

  /**
   * Finds the current position in the active members array
   */
  private findCurrentPositionInActive(activeMembers: Member[]): number {
    if (activeMembers.length === 0) return 0;

    // Find where the current member is in the active list
    const currentMember = this.members[this.currentIndex];
    let position = activeMembers.findIndex((m) => m.id === currentMember.id);

    // If current member is inactive or not found, start from beginning
    if (position === -1) {
      position = 0;
    }
    // Note: We don't advance here - we'll use this position and advance after selection

    return position;
  }

  /**
   * Finds the index of a member in the original members array
   */
  private findMemberIndexInOriginal(memberId: number): number {
    const index = this.members.findIndex((m) => m.id === memberId);
    return index >= 0 ? index : this.currentIndex;
  }

  /**
   * Resets the iterator to the beginning
   */
  reset(): void {
    this.currentIndex = 0;
    this.hasStarted = false;
  }
}
