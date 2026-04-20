import { TeamRotator } from './TeamRotator';
import { createMember, Member } from './models/member';

describe('TeamRotator - Iterator Pattern Implementation', () => {
  /**
   * Example 1: Basic Rotation
   * Tests that members rotate in round-robin order
   */
  describe('Example 1: Basic Rotation', () => {
    it('should rotate through 4 active members in round-robin order', () => {
      // Arrange
      const members: Member[] = [
        createMember(1, 'Alice'),
        createMember(2, 'Bob'),
        createMember(3, 'Charlie'),
        createMember(4, 'Diana'),
      ];
      const rotator = new TeamRotator(members);

      // Act & Assert
      expect(rotator.next()?.name).toBe('Alice');
      expect(rotator.next()?.name).toBe('Bob');
      expect(rotator.next()?.name).toBe('Charlie');
      expect(rotator.next()?.name).toBe('Diana');
      expect(rotator.next()?.name).toBe('Alice'); // Rotation restarts
      expect(rotator.next()?.name).toBe('Bob');
    });
  });

  /**
   * Example 2: No Immediate Repetition
   * Scenario: Alice was just selected manually or externally
   */
  describe('Example 2: No Immediate Repetition', () => {
    it('should not return Alice immediately when she was last selected externally', () => {
      // Arrange - Alice (id: 1) was last selected
      const members: Member[] = [
        createMember(1, 'Alice'),
        createMember(2, 'Bob'),
        createMember(3, 'Charlie'),
      ];
      const rotator = new TeamRotator(members, 1); // Alice was last selected

      // Act & Assert
      expect(rotator.next()?.name).toBe('Bob'); // NOT Alice, because she was last
      expect(rotator.next()?.name).toBe('Charlie');
      expect(rotator.next()?.name).toBe('Alice'); // Now OK to return Alice
      expect(rotator.next()?.name).toBe('Bob');
    });

    it('should not return the same member twice in a row during normal operation', () => {
      // Arrange
      const members: Member[] = [
        createMember(1, 'Alice'),
        createMember(2, 'Bob'),
        createMember(3, 'Charlie'),
      ];
      const rotator = new TeamRotator(members);

      // Act & Assert
      expect(rotator.next()?.name).toBe('Alice');
      expect(rotator.next()?.name).toBe('Bob'); // NOT Alice
      expect(rotator.next()?.name).toBe('Charlie');
      expect(rotator.next()?.name).toBe('Alice'); // Now OK to return Alice
      expect(rotator.next()?.name).toBe('Bob');
    });

    it('should not return Bob immediately when he was last selected externally, and should return the next member in the list to ensure fair rotation', () => {
      // Arrange - Alice (id: 1) was last selected
      const members: Member[] = [
        createMember(1, 'Alice'),
        createMember(2, 'Bob'),
        createMember(3, 'Charlie'),
      ];
      const rotator = new TeamRotator(members, 2); // Bob was last selected

      // Act & Assert
      expect(rotator.next()?.name).toBe('Charlie'); // NOT Bob, because he was last selected externally, and we start from Charlie to ensure fair rotation
      expect(rotator.next()?.name).toBe('Alice'); // Now OK to return Alice
      expect(rotator.next()?.name).toBe('Bob');
    });
  });

  /**
   * Example 3: Skipping Inactive Members
   * Tests that inactive members are automatically skipped
   */
  describe('Example 3: Skipping Inactive Members', () => {
    it('should skip Bob who is inactive and rotate through others', () => {
      // Arrange
      const members: Member[] = [
        createMember(1, 'Alice', true),
        createMember(2, 'Bob', false), // Inactive
        createMember(3, 'Charlie', true),
        createMember(4, 'Diana', true),
      ];
      const rotator = new TeamRotator(members);

      // Act & Assert
      expect(rotator.next()?.name).toBe('Alice');
      expect(rotator.next()?.name).toBe('Charlie'); // Skips Bob
      expect(rotator.next()?.name).toBe('Diana');
      expect(rotator.next()?.name).toBe('Alice'); // Skips Bob
      expect(rotator.next()?.name).toBe('Charlie');
    });
  });

  /**
   * Example 4: Requesting Next N Members
   * Tests that multiple members can be requested at once
   */
  describe('Example 4: Requesting Next N Members', () => {
    it('should return next 2 members at a time in correct rotation', () => {
      // Arrange
      const members: Member[] = [
        createMember(1, 'Alice'),
        createMember(2, 'Bob'),
        createMember(3, 'Charlie'),
        createMember(4, 'Diana'),
      ];
      const rotator = new TeamRotator(members);

      // Act & Assert
      const batch1 = rotator.nextN(2);
      expect(batch1.map((m) => m.name)).toEqual(['Alice', 'Bob']);

      const batch2 = rotator.nextN(2);
      expect(batch2.map((m) => m.name)).toEqual(['Charlie', 'Diana']);

      const batch3 = rotator.nextN(2);
      expect(batch3.map((m) => m.name)).toEqual(['Alice', 'Bob']); // Rotation restarts
    });

    it('should throw error when requesting more members than available', () => {
      // Arrange
      const members: Member[] = [createMember(1, 'Alice'), createMember(2, 'Bob')];
      const rotator = new TeamRotator(members);

      // Act & Assert
      expect(() => rotator.nextN(5)).toThrow(
        'Cannot select 5 members: only 2 active members available',
      );
    });
  });

  /**
   * Example 5: Edge Case - Only One Active Member
   * Tests that repetition is allowed when only one member is active
   */
  describe('Example 5: Edge Case - Only One Active Member', () => {
    it('should return Alice repeatedly when she is the only active member', () => {
      // Arrange
      const members: Member[] = [
        createMember(1, 'Alice', true),
        createMember(2, 'Bob', false),
        createMember(3, 'Charlie', false),
      ];
      const rotator = new TeamRotator(members);

      // Act & Assert - Alice should be returned every time
      expect(rotator.next()?.name).toBe('Alice');
      expect(rotator.next()?.name).toBe('Alice'); // Repetition is acceptable
      expect(rotator.next()?.name).toBe('Alice');
    });
  });

  /**
   * Example 6: Edge Case - All Members Inactive
   * Tests error handling when no members are available
   */
  describe('Example 6: Edge Case - All Members Inactive', () => {
    it('should return null when all members are inactive', () => {
      // Arrange
      const members: Member[] = [
        createMember(1, 'Alice', false),
        createMember(2, 'Bob', false),
        createMember(3, 'Charlie', false),
      ];
      const rotator = new TeamRotator(members);

      // Act & Assert
      expect(rotator.next()).toBeNull();
      expect(rotator.hasNext()).toBe(false);
    });
  });
});
