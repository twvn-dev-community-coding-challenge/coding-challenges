/**
 * Example 6: Edge Case - All Members Inactive
 *
 * Scenario: Everyone is on vacation or unavailable.
 * Expected: Error or empty result with message "No active members available".
 */

import { TeamRotator } from '@team-rotator/core';
import type { Member } from '@team-rotator/core';

const members: Member[] = [
  { id: 1, name: 'Alice', isActive: false },
  { id: 2, name: 'Bob', isActive: false },
  { id: 3, name: 'Charlie', isActive: false },
];

console.log('🚀 Example 6: Edge Case - All Members Inactive ===\n');
console.log('Team Setup:');
members.forEach((m) => console.log(`- ${m.name} (id: ${m.id}, inactive)`));
console.log('\nRotation Results:');


try {
  new TeamRotator(members);
  console.log('ERROR: Should have thrown an error!');
} catch (error) {
  const err = error as Error;
  console.log(`Call getNext() → Error: ${err.message}`);
}

console.log('\n✅ Demonstrates handling the case when no one is available.\n');
