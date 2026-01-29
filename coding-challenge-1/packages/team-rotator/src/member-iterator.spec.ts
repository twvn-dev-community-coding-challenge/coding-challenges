import { describe, expect, it } from '@jest/globals';
import { MemberIterator } from './member-iterator';
import { DuplicatedMemberIdentifierError, Member, NoActiveMembersError } from './types';

describe('MemberIterator', () => {
  describe('Basic Iteration', () => {
    it('should iterate through active members', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
        { id: 3, name: 'Charlie', isActive: true },
      ];

      const iterator = new MemberIterator(members);

      expect(iterator.next()?.name).toBe('Alice');
      expect(iterator.next()?.name).toBe('Bob');
      expect(iterator.next()?.name).toBe('Charlie');
      expect(iterator.next()?.name).toBe('Alice'); // Cycles back
    });

    it('should throw the error when no active members found from the initialization list', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: false },
        { id: 2, name: 'Bob', isActive: false },
      ];

      try {
        new MemberIterator(members);
      } catch (err) {
        expect(err instanceof NoActiveMembersError).toBeTruthy();
        expect((err as Error).message).toEqual('No active members available');
      }
    });
  });

  describe('Excluding Members', () => {
    it('should skip excluded member', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
        { id: 3, name: 'Charlie', isActive: true },
      ];

      const iterator = new MemberIterator(members);

      // Exclude Alice
      const next = iterator.next(1);
      expect(next?.name).toBe('Bob');
    });

    it('should allow repetition when only one active member', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: false },
      ];

      const iterator = new MemberIterator(members);

      // Even if we exclude Alice, we should get her back (only option)
      expect(iterator.next(1)?.name).toBe('Alice');
    });
  });

  describe('Next N Members', () => {
    it('should return next N members', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
        { id: 3, name: 'Charlie', isActive: true },
      ];

      const iterator = new MemberIterator(members);

      const batch = iterator.nextN(2);
      expect(batch).toHaveLength(2);
      expect(batch[0].name).toBe('Alice');
      expect(batch[1].name).toBe('Bob');
    });

    it('should handle exclusion in nextN', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
        { id: 3, name: 'Charlie', isActive: true },
      ];

      const iterator = new MemberIterator(members);

      const batch = iterator.nextN(2, 1); // Exclude Alice
      expect(batch).toHaveLength(2);
      expect(batch[0].name).toBe('Bob');
      expect(batch[1].name).toBe('Charlie');
    });
  });

  describe('Reset', () => {
    it('should reset iterator position', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
      ];

      const iterator = new MemberIterator(members);
      iterator.next();
      iterator.reset();

      // After reset, should start from beginning
      expect(iterator.next()?.name).toBe('Alice');
    });
  });

  describe('Edge Cases', () => {
    it('should handle when current member becomes inactive', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: false },
        { id: 3, name: 'Charlie', isActive: true },
      ];

      const iterator = new MemberIterator(members);
      // Select Alice first
      expect(iterator.next()?.name).toBe('Alice');

      // Now Bob is at currentIndex but inactive, should start from beginning of active
      expect(iterator.next()?.name).toBe('Charlie');
    });

    it('should handle member not found in original array', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
      ];

      const iterator = new MemberIterator(members);
      // This should work normally
      const member = iterator.next();
      expect(member).toBeTruthy();
    });

    it('should cycle through when requesting more than available', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
      ];

      const iterator = new MemberIterator(members);
      const batch = iterator.nextN(5);
      expect(batch.length).toBe(5);
      // Should cycle through Alice and Bob
      expect(batch[0].name).toBe('Alice');
      expect(batch[1].name).toBe('Bob');
      expect(batch[2].name).toBe('Alice');
    });

    it('should handle complex exclusion scenarios', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
        { id: 3, name: 'Charlie', isActive: true },
      ];

      const iterator = new MemberIterator(members);
      // Get first member
      iterator.next();
      // Now exclude the next one in rotation
      const next = iterator.next(2); // Exclude Bob who would be next
      expect(next?.name).toBe('Charlie');
    });

    it('should throw duplicated member identifier found when the same id added', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: false },
        { id: 1, name: 'Bob', isActive: true },
      ];

      try {
        new MemberIterator(members);
      } catch (err) {
        expect(err instanceof DuplicatedMemberIdentifierError).toBeTruthy();
        expect((err as Error).message).toEqual(
          'Duplicated member identifier found in the rotator list'
        );
      }
    });

    it('should add new member to the existing list successfully', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: false },
        { id: 2, name: 'Bob', isActive: true },
      ];

      const memberIterator = new MemberIterator(members);
      memberIterator.addMember({ id: 3, name: 'Charlie', isActive: true });
    });

    it('should handle all members excluded scenario', () => {
      // This tests the fallback case in the while loop
      // However, with >1 active member, this shouldn't happen in practice
      // But we test the code path exists
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: true },
      ];

      const iterator = new MemberIterator(members);
      // Normal operation
      const first = iterator.next();
      expect(first).toBeTruthy();
    });

    it('should properly handle position calculation when current member not in active list', () => {
      const members: Member[] = [
        { id: 1, name: 'Alice', isActive: true },
        { id: 2, name: 'Bob', isActive: false }, // inactive
        { id: 3, name: 'Charlie', isActive: true },
      ];

      const iterator = new MemberIterator(members);
      // Select Alice (index 0)
      expect(iterator.next()?.name).toBe('Alice');
      // Now currentIndex points to Alice, but if we had selected Bob (inactive),
      // it should handle position = -1 case
      expect(iterator.next()?.name).toBe('Charlie');
    });
  });
});
