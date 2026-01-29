import { describe, expect, it } from '@jest/globals';
import { TeamRotator } from './team-rotator';
import { Member, NoActiveMembersError } from './types';

describe('TeamRotator', () => {
  describe('Basic Rotation', () => {
    it('should rotate through all members in order', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
        { id: 3, name: 'Charlie', isActive: true },
        { id: 4, name: 'Diana', isActive: true },
      ];

      const rotator = new TeamRotator(members);

      expect(rotator.getNext().name).toBe('Alice');
      expect(rotator.getNext().name).toBe('Bob');
      expect(rotator.getNext().name).toBe('Charlie');
      expect(rotator.getNext().name).toBe('Diana');
      expect(rotator.getNext().name).toBe('Alice'); // Rotation restarts
      expect(rotator.getNext().name).toBe('Bob');
    });

    it('should track last selected member', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
      ];

      const rotator = new TeamRotator(members);
      expect(rotator.getLastSelectedMemberId()).toBeNull();

      rotator.getNext();
      expect(rotator.getLastSelectedMemberId()).toBe(1);

      rotator.getNext();
      expect(rotator.getLastSelectedMemberId()).toBe(2);
    });
  });

  describe('No Immediate Repetition', () => {
    it('should not return the same member twice in a row', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
        { id: 3, name: 'Charlie', isActive: true },
      ];

      const rotator = new TeamRotator(members);

      // Manually set last selected to Alice
      rotator.setLastSelectedMember(1);

      // Next call should return Bob, not Alice
      expect(rotator.getNext().name).toBe('Bob');
      expect(rotator.getNext().name).toBe('Charlie');
      expect(rotator.getNext().name).toBe('Alice'); // Now OK to return Alice
    });

    it('should allow repetition when only one active member exists', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: false },
        { id: 3, name: 'Charlie', isActive: false },
      ];

      const rotator = new TeamRotator(members);

      expect(rotator.getNext().name).toBe('Alice');
      expect(rotator.getNext().name).toBe('Alice'); // Repetition allowed
      expect(rotator.getNext().name).toBe('Alice');
    });
  });

  describe('Skipping Inactive Members', () => {
    it('should skip inactive members during rotation', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: false },
        { id: 3, name: 'Charlie', isActive: true },
        { id: 4, name: 'Diana', isActive: true },
      ];

      const rotator = new TeamRotator(members);

      expect(rotator.getNext().name).toBe('Alice');
      expect(rotator.getNext().name).toBe('Charlie'); // Skips Bob
      expect(rotator.getNext().name).toBe('Diana');
      expect(rotator.getNext().name).toBe('Alice'); // Skips Bob again
      expect(rotator.getNext().name).toBe('Charlie');
    });

    it('should throw error when all members are inactive', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: false },
        { id: 2, name: 'Bob', isActive: false },
        { id: 3, name: 'Charlie', isActive: false },
      ];
      try {
        new TeamRotator(members);
      } catch (err) {
        expect(err instanceof NoActiveMembersError).toBeTruthy();
        expect((err as Error).message).toEqual('No active members available');
      }
    });
  });

  describe('Getting Next N Members', () => {
    it('should return next N members', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
        { id: 3, name: 'Charlie', isActive: true },
        { id: 4, name: 'Diana', isActive: true },
      ];

      const rotator = new TeamRotator(members);

      const firstBatch = rotator.getNextN(2);
      expect(firstBatch).toHaveLength(2);
      expect(firstBatch[0].name).toBe('Alice');
      expect(firstBatch[1].name).toBe('Bob');

      const secondBatch = rotator.getNextN(2);
      expect(secondBatch).toHaveLength(2);
      expect(secondBatch[0].name).toBe('Charlie');
      expect(secondBatch[1].name).toBe('Diana');

      const thirdBatch = rotator.getNextN(2);
      expect(thirdBatch).toHaveLength(2);
      expect(thirdBatch[0].name).toBe('Alice'); // Rotation restarts
      expect(thirdBatch[1].name).toBe('Bob');
    });

    it('should handle getting more members than available', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
      ];

      const rotator = new TeamRotator(members);

      const batch = rotator.getNextN(5);
      expect(batch.length).toBeGreaterThan(0);
      // Should cycle through available members
    });

    it('should throw error for invalid count', () => {
      const members: Member[] = [{ id: 1, name: 'Alice', isActive: true }];

      const rotator = new TeamRotator(members);

      expect(() => rotator.getNextN(0)).toThrow('Count must be greater than 0');
      expect(() => rotator.getNextN(-1)).toThrow('Count must be greater than 0');
    });
  });

  describe('Edge Cases', () => {
    it('should handle single active member', () => {
      const members: Member[] = [{ id: 1, name: 'Alice', isActive: true }];

      const rotator = new TeamRotator(members);

      expect(rotator.getNext().name).toBe('Alice');
      expect(rotator.getNext().name).toBe('Alice');
    });

    it('should throw error when team is empty', () => {
      expect(() => new TeamRotator([])).toThrow('Team must have at least one member');
    });

    it('should allow setting last selected member manually', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
      ];

      const rotator = new TeamRotator(members);
      rotator.setLastSelectedMember(1);

      expect(rotator.getLastSelectedMemberId()).toBe(1);
      expect(rotator.getNext().name).toBe('Bob'); // Should skip Alice
    });

    it('should throw error when setting invalid member ID', () => {
      const members: Member[] = [{ id: 1, name: 'Alice', isActive: true }];

      const rotator = new TeamRotator(members);

      expect(() => rotator.setLastSelectedMember(999)).toThrow('Member with id 999 not found');
    });

    it('should reset rotation state', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
      ];

      const rotator = new TeamRotator(members);
      rotator.getNext();
      expect(rotator.getLastSelectedMemberId()).toBe(1);

      rotator.reset();
      expect(rotator.getLastSelectedMemberId()).toBeNull();
    });

    it('should return copy of members', () => {
      const members: Member[] = [{ id: 1, name: 'Alice', isActive: true }];

      const rotator = new TeamRotator(members);
      const returnedMembers = rotator.getMembers();

      // Modifying returned array should not affect internal state
      returnedMembers.push({ id: 2, name: 'Bob', isActive: true });
      expect(rotator.getMembers()).toHaveLength(1);
    });
  });

  describe('Fair Rotation', () => {
    it('should ensure fair distribution over multiple rotations', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
        { id: 3, name: 'Charlie', isActive: true },
      ];

      const rotator = new TeamRotator(members);
      const selections: string[] = [];

      // Perform 9 rotations (3 full cycles)
      for (let i = 0; i < 9; i++) {
        selections.push(rotator.getNext().name);
      }

      // Count occurrences
      const aliceCount = selections.filter((n) => n === 'Alice').length;
      const bobCount = selections.filter((n) => n === 'Bob').length;
      const charlieCount = selections.filter((n) => n === 'Charlie').length;

      // Each member should be selected exactly 3 times (fair distribution)
      expect(aliceCount).toBe(3);
      expect(bobCount).toBe(3);
      expect(charlieCount).toBe(3);
    });
  });
});
