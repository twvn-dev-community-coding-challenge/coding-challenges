/**
 * Example 5: Edge Case - Only One Active Member
 *
 * Scenario: Everyone except Alice is inactive.
 * Expected: Alice can be returned multiple times (repetition acceptable).
 */

import { TeamRotator } from '@team-rotator/core';
import type { Member } from '@team-rotator/core';

const members: Member[] = [
  { id: 1, name: 'Alice', isActive: true }, // only active member
  { id: 2, name: 'Bob', isActive: false },
  { id: 3, name: 'Charlie', isActive: false },
];

console.log('🚀 Example 5: Edge Case - Only One Active Member ===\n');
console.log('Team Setup:');
members.forEach((m) => {
  const status = m.isActive ? 'active ← only active member' : 'inactive';
  console.log(`- ${m.name} (id: ${m.id}, ${status})`);
});
console.log('\nRotation Results:');

const rotator = new TeamRotator(members);

for (let i = 1; i <= 3; i++) {
  const member = rotator.getNext();
  console.log(`${i}st call → Returns: ${member.name}`);
}

console.log('\n✅ Demonstrates that repetition is acceptable when only one active member.\n');
