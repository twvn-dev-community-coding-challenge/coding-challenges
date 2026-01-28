/**
 * Represents a team member in the rotation system
 */
export interface Member {
  id: number;
  name: string;
  isActive: boolean;
}

/**
 * Factory function to create a Member
 */
export function createMember(id: number, name: string, isActive: boolean = true): Member {
  return { id, name, isActive };
}