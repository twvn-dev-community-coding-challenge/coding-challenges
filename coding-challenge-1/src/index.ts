import { TeamRotator } from './TeamRotator';
import { createMember } from './models/member';

/**
 * Example usage of the TeamRotator API
 */
function main() {
  console.log('=== Smart Team Rotator API - Example Usage ===\n');

  // Example 1: Basic rotation
  console.log('Example 1: Basic rotation');
  const team = [
    createMember(1, 'Alice'),
    createMember(2, 'Bob'),
    createMember(3, 'Charlie'),
    createMember(4, 'Diana')
  ];

  const rotator = new TeamRotator(team);

  console.log('Next member:', rotator.next()?.name); // Alice
  console.log('Next member:', rotator.next()?.name); // Bob
  console.log('Next member:', rotator.next()?.name); // Charlie
  console.log('Next member:', rotator.next()?.name); // Diana
  console.log('Next member:', rotator.next()?.name); // Alice (rotation restarts)
  console.log();

  // Example 2: No immediate repetition (Alice was selected externally)
  console.log('Example 2: No immediate repetition');
  const team2 = [
    createMember(1, 'Alice'),
    createMember(2, 'Bob'),
    createMember(3, 'Charlie')
  ];

  // Alice (id: 1) was just selected manually/externally
  const rotator2 = new TeamRotator(team2, 1);
  console.log('Next member:', rotator2.next()?.name); // Bob (NOT Alice)
  console.log('Next member:', rotator2.next()?.name); // Charlie
  console.log('Next member:', rotator2.next()?.name); // Alice (now OK)
  console.log('Next member:', rotator2.next()?.name); // Bob
  console.log();

  // Example 3: Get next N members
  console.log('Example 3: Request multiple members at once');
  rotator.reset();
  const batch1 = rotator.nextN(2);
  console.log('First batch:', batch1.map(m => m.name)); // [Alice, Bob]

  const batch2 = rotator.nextN(2);
  console.log('Second batch:', batch2.map(m => m.name)); // [Charlie, Diana]
  console.log();

  // Example 4: Team with inactive members
  console.log('Example 4: Skipping inactive members');
  const teamWithInactive = [
    createMember(1, 'Alice', true),
    createMember(2, 'Bob', false), // Inactive
    createMember(3, 'Charlie', true),
    createMember(4, 'Diana', true)
  ];

  const rotator3 = new TeamRotator(teamWithInactive);
  console.log('Next member:', rotator3.next()?.name); // Alice
  console.log('Next member:', rotator3.next()?.name); // Charlie (skips Bob)
  console.log('Next member:', rotator3.next()?.name); // Diana
  console.log('Next member:', rotator3.next()?.name); // Alice (skips Bob)
  console.log();

  // Example 5: Edge case - only one active member
  console.log('Example 5: Only one active member');
  const teamWithOneActive = [
    createMember(1, 'Alice', true),
    createMember(2, 'Bob', false),
    createMember(3, 'Charlie', false)
  ];

  const rotator4 = new TeamRotator(teamWithOneActive);
  console.log('Next member:', rotator4.next()?.name); // Alice
  console.log('Next member:', rotator4.next()?.name); // Alice (repetition allowed)
  console.log('Next member:', rotator4.next()?.name); // Alice
  console.log();

  // Example 6: Edge case - all inactive
  console.log('Example 6: All members inactive');
  const teamAllInactive = [
    createMember(1, 'Alice', false),
    createMember(2, 'Bob', false)
  ];

  const rotator5 = new TeamRotator(teamAllInactive);
  console.log('Has next?:', rotator5.hasNext()); // false
  console.log('Next member:', rotator5.next()); // null
  console.log();

  console.log('=== Examples completed ===');
}

// Run examples if this file is executed directly
if (require.main === module) {
  main();
}

export { main };