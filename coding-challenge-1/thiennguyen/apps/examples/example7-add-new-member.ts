/**
 * Example 6: Edge Case - All Members Inactive
 *
 * Scenario: Everyone is on vacation or unavailable.
 * Expected: Error or empty result with message "No active members available".
 */

import { TeamRotator } from '@team-rotator/core';
import type { Member } from '@team-rotator/core';

const members: Member[] = [
  { id: 1, name: 'Alice', isActive: true },
  { id: 2, name: 'Bob', isActive: false },
  { id: 3, name: 'Charlie', isActive: false },
];

console.log('🚀 Example 7: Add new member to the rotator list ===\n');
console.log('Team Setup:');
members.forEach((m) => {
  const status = m.isActive ? 'active' : 'inactive ← unavailable';
  console.log(`- ${m.name} (id: ${m.id}, ${status})`);
});
console.log('\nRotation Results:');

const rotator = new TeamRotator(members);

for (let i = 1; i <= 3; i++) {
  const member = rotator.getNext();
  console.log(`${i}st call → Returns: ${member.name}`);
}
const newMember = { id: 4, name: 'David', isActive: true };
console.log(`\nAdd new member - ${newMember.name} (id: ${newMember.id}, active)`);

rotator.addNewMember(newMember);
for (let i = 1; i <= 4; i++) {
  const member = rotator.getNext();
  console.log(`${i + 3}st call → Returns: ${member.name}`);
}

console.log(
  '\n✅ Demonstrates handling the case when add new member into the current rotator list.\n'
);
