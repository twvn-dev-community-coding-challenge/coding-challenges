import { Member } from '../models/member';

/**
 * Iterator Pattern Interface
 * Provides a way to sequentially access team members without exposing the underlying collection
 */
export interface RotationIterator {
  /**
   * Returns the next member in rotation
   * @returns The next active member, or null if no active members available
   */
  next(): Member | null;

  /**
   * Returns the next N members in rotation
   * @param count Number of members to return
   * @returns Array of next active members (may be less than count if not enough active members)
   */
  nextN(count: number): Member[];

  /**
   * Checks if there are any active members available
   * @returns true if at least one active member exists
   */
  hasNext(): boolean;

  /**
   * Resets the iterator state (useful for testing)
   */
  reset(): void;
}