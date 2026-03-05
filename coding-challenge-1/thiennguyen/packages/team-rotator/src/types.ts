/**
 * Represents a team member
 */
export interface Member {
  id: number;
  name: string;
  isActive: boolean;
}

/**
 * Error thrown when no active members are available
 */
export class NoActiveMembersError extends Error {
  constructor(message: string = 'No active members available') {
    super(message);
    this.name = 'NoActiveMembersError';
  }
}

/**
 * Error thrown when duplicated member id found
 */
export class DuplicatedMemberIdentifierError extends Error {
  constructor(message: string = 'Duplicated member identifier found in the rotator list') {
    super(message);
    this.name = 'DuplicatedMemberIdentifierError';
  }
}
