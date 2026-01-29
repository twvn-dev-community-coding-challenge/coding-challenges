/**
 * Example 1: Basic Rotation
 *
 * Scenario: Your team has 4 active members.
 * Expected: Simple round-robin rotation through all members.
 */

import { TeamRotator } from '@team-rotator/core';
import type { Member } from '@team-rotator/core';

const members: Member[] = [
  { id: 1, name: 'Alice', isActive: true },
  { id: 2, name: 'Bob', isActive: true },
  { id: 3, name: 'Charlie', isActive: true },
  { id: 4, name: 'Diana', isActive: true },
];

console.log('🚀 Example 1: Basic Rotation ===\n');
console.log('Team Setup:');
members.forEach((m) => console.log(`- ${m.name} (id: ${m.id}, active: ${m.isActive})`));
console.log('\nRotation Results:');

const rotator = new TeamRotator(members);

for (let i = 1; i <= 6; i++) {
  const member = rotator.getNext();
  console.log(`${i}st call → Returns: ${member.name}`);
}

console.log('\n✅ Demonstrates simple round-robin rotation.\n');
