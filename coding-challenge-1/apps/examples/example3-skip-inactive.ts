/**
 * Example 3: Skipping Inactive Members
 *
 * Scenario: Bob becomes inactive in the middle of rotation.
 * Expected: Bob should be automatically skipped.
 */

import { TeamRotator } from '@team-rotator/core';
import type { Member } from '@team-rotator/core';

const members: Member[] = [
  { id: 1, name: 'Alice', isActive: true },
  { id: 2, name: 'Bob', isActive: false }, // inactive
  { id: 3, name: 'Charlie', isActive: true },
  { id: 4, name: 'Diana', isActive: true },
];

console.log('=== Example 3: Skipping Inactive Members ===\n');
console.log('Team Setup:');
members.forEach((m) => {
  const status = m.isActive ? 'active' : 'inactive ← unavailable';
  console.log(`- ${m.name} (id: ${m.id}, ${status})`);
});
console.log('\nRotation Results:');

const rotator = new TeamRotator(members);

for (let i = 1; i <= 5; i++) {
  const member = rotator.getNext();
  console.log(`${i}st call → Returns: ${member.name}`);
}

console.log('\n✅ Demonstrates skipping inactive members automatically.\n');
