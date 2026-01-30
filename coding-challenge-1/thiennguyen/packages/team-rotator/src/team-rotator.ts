import { Member, NoActiveMembersError } from './types';
import { MemberIterator } from './member-iterator';

/**
 * Team Rotator - Main API
 *
 * Manages team rotation with the following features:
 * - Round-robin rotation through active members
 * - No immediate repetition
 * - Fair distribution over time
 * - Support for getting next member or next N members
 *
 * Uses Iterator Pattern for cycling through members
 */
export class TeamRotator {
  private members: Member[];
  private iterator: MemberIterator;
  private lastSelectedMemberId: number | null = null;

  /**
   * Creates a new TeamRotator instance
   * @param members - Array of team members
   */
  constructor(members: Member[]) {
    if (!members || members.length === 0) {
      throw new Error('Team must have at least one member');
    }
    this.members = [...members]; // Create a copy to avoid external mutations
    this.iterator = new MemberIterator(this.members);
  }

  /**
   * Gets the next member for rotation
   * @returns The next member
   * @throws NoActiveMembersError if no active members are available
   */
  getNext(): Member {
    const member = this.iterator.next(this.lastSelectedMemberId ?? undefined);

    if (!member) {
      throw new NoActiveMembersError();
    }

    this.lastSelectedMemberId = member.id;
    return member;
  }

  /**
   * Gets the next N members for rotation
   * @param count - Number of members to return
   * @returns Array of next N members
   * @throws NoActiveMembersError if no active members are available
   */
  getNextN(count: number): Member[] {
    if (count <= 0) {
      throw new Error('Count must be greater than 0');
    }

    const members = this.iterator.nextN(count, this.lastSelectedMemberId ?? undefined);

    if (members.length === 0) {
      throw new NoActiveMembersError();
    }

    // Update last selected to the last member in the batch
    if (members.length > 0) {
      this.lastSelectedMemberId = members[members.length - 1].id;
    }

    return members;
  }

  /**
   * Gets all team members
   * @returns Copy of all members
   */
  getMembers(): Member[] {
    return [...this.members];
  }

  /**
   * Add new member to the current rotator list (new member will come last)
   * @param member
   * @throws DuplicatedMemberIdentifierError if duplicated member id found from the list
   */
  addNewMember(member: Member): void {
    this.iterator.addMember(member);
  }

  /**
   * Gets the last selected member ID
   * @returns Last selected member ID or null
   */
  getLastSelectedMemberId(): number | null {
    return this.lastSelectedMemberId;
  }

  /**
   * Manually sets the last selected member (useful for external selections)
   * @param memberId - ID of the member that was selected externally
   */
  setLastSelectedMember(memberId: number): void {
    const member = this.members.find((m) => m.id === memberId);
    if (!member) {
      throw new Error(`Member with id ${memberId} not found`);
    }
    this.lastSelectedMemberId = memberId;
  }

  /**
   * Resets the rotation state
   */
  reset(): void {
    this.lastSelectedMemberId = null;
    this.iterator.reset();
  }
}
