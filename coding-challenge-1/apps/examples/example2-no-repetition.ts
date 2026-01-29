/**
 * Example 2: No Immediate Repetition
 *
 * Scenario: Alice was just selected manually or externally.
 * Expected: Next call should return Bob (NOT Alice).
 */

import { TeamRotator } from '@team-rotator/core';
import type { Member } from '@team-rotator/core';

const members: Member[] = [
  { id: 1, name: 'Alice', isActive: true },
  { id: 2, name: 'Bob', isActive: true },
  { id: 3, name: 'Charlie', isActive: true },
];

console.log('🚀 Example 2: No Immediate Repetition ===\n');
console.log('Team Setup:');
members.forEach((m) => console.log(`- ${m.name} (id: ${m.id}, active: ${m.isActive})`));
console.log('- Alice ← last selected\n');

const rotator = new TeamRotator(members);
rotator.setLastSelectedMember(1); // Manually set Alice as last selected

console.log('Rotation Results:');
for (let i = 1; i <= 4; i++) {
  const member = rotator.getNext();
  console.log(`${i}st call → Returns: ${member.name}`);
}

console.log('\n✅ Demonstrates the "no immediate repetition" rule.\n');
