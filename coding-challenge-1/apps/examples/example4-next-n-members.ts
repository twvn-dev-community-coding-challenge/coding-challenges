/**
 * Example 4: Requesting Next N Members
 *
 * Scenario: You need 2 people for a task.
 * Expected: Returns array of next N members.
 */

import { TeamRotator } from '@team-rotator/core';
import type { Member } from '@team-rotator/core';

const members: Member[] = [
  { id: 1, name: 'Alice', isActive: true },
  { id: 2, name: 'Bob', isActive: true },
  { id: 3, name: 'Charlie', isActive: true },
  { id: 4, name: 'Diana', isActive: true },
];

console.log('=== Example 4: Requesting Next N Members ===\n');
console.log('Team Setup:');
members.forEach((m) => console.log(`- ${m.name} (id: ${m.id}, active: ${m.isActive})`));
console.log('\nRotation Results:');

const rotator = new TeamRotator(members);

for (let i = 1; i <= 3; i++) {
  const batch = rotator.getNextN(2);
  const names = batch.map((m) => m.name).join(', ');
  console.log(`Call getNext(n=2) → Returns: [${names}]`);
}

console.log('\n✅ Demonstrates returning multiple members at once.\n');
