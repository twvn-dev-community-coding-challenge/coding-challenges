#!/usr/bin/env node

import * as readline from 'readline';
import { TeamRotator } from './TeamRotator';
import { createMember, Member } from './models/Member';

// CLI state
let rotator: TeamRotator | null = null;
let team: Member[] = [];

// Create readline interface
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  prompt: 'team-rotator> '
});

// Display welcome message
function displayWelcome() {
  console.log('\n=== Smart Team Rotator CLI ===\n');
  console.log('Commands:');
  console.log('  init <members>        - Initialize team (e.g., init Alice,Bob,Charlie)');
  console.log('  init-with <members> <lastId> - Initialize with last selected ID');
  console.log('  add <name> [active]   - Add member (active=true by default)');
  console.log('  next                  - Get next member');
  console.log('  next <n>              - Get next N members');
  console.log('  show                  - Show current team');
  console.log('  status                - Show rotator status');
  console.log('  reset                 - Reset rotation state');
  console.log('  help                  - Show this help');
  console.log('  exit                  - Exit the CLI');
  console.log();
}

// Display help
function displayHelp() {
  console.log('\nAvailable Commands:\n');
  console.log('  init <members>');
  console.log('    Initialize team with comma-separated names');
  console.log('    Example: init Alice,Bob,Charlie,Diana\n');

  console.log('  init-with <members> <lastId>');
  console.log('    Initialize team with a pre-selected member');
  console.log('    Example: init-with Alice,Bob,Charlie 1\n');

  console.log('  add <name> [active]');
  console.log('    Add a member to the team (active=true/false, default=true)');
  console.log('    Example: add Alice');
  console.log('    Example: add Bob false\n');

  console.log('  next');
  console.log('    Get the next member in rotation\n');

  console.log('  next <n>');
  console.log('    Get the next N members in rotation');
  console.log('    Example: next 3\n');

  console.log('  show');
  console.log('    Display all team members and their status\n');

  console.log('  status');
  console.log('    Show current rotator state\n');

  console.log('  reset');
  console.log('    Reset rotation to beginning\n');
}

// Initialize team from comma-separated names
function initTeam(names: string[], lastSelectedId?: number) {
  team = names.map((name, index) => createMember(index + 1, name.trim()));
  rotator = new TeamRotator(team, lastSelectedId);
  console.log(`✓ Team initialized with ${team.length} members`);
  if (lastSelectedId) {
    const lastMember = team.find(m => m.id === lastSelectedId);
    console.log(`  Last selected: ${lastMember?.name || 'Unknown'} (id: ${lastSelectedId})`);
  }
}

// Add a member to the team
function addMember(name: string, isActive: boolean = true) {
  if (!rotator) {
    console.log('⚠ Please initialize the team first using "init" command');
    return;
  }

  const newId = team.length + 1;
  const member = createMember(newId, name, isActive);
  team.push(member);

  // Reinitialize rotator with updated team
  rotator = new TeamRotator(team);
  console.log(`✓ Added ${name} (id: ${newId}, active: ${isActive})`);
}

// Show team members
function showTeam() {
  if (team.length === 0) {
    console.log('⚠ No team members. Use "init" to create a team.');
    return;
  }

  console.log('\nTeam Members:');
  console.log('─'.repeat(50));
  team.forEach(member => {
    const status = member.isActive ? '✓ Active' : '✗ Inactive';
    console.log(`  ${member.id}. ${member.name.padEnd(20)} ${status}`);
  });
  console.log('─'.repeat(50));
  console.log();
}

// Show rotator status
function showStatus() {
  if (!rotator) {
    console.log('⚠ Rotator not initialized');
    return;
  }

  console.log('\nRotator Status:');
  console.log('─'.repeat(50));
  console.log(`  Total members: ${team.length}`);
  console.log(`  Active members: ${team.filter(m => m.isActive).length}`);
  console.log(`  Has next: ${rotator.hasNext()}`);
  console.log('─'.repeat(50));
  console.log();
}

// Get next member
function getNext() {
  if (!rotator) {
    console.log('⚠ Please initialize the team first using "init" command');
    return;
  }

  const next = rotator.next();
  if (next) {
    console.log(`→ Next member: ${next.name} (id: ${next.id})`);
  } else {
    console.log('⚠ No active members available');
  }
}

// Get next N members
function getNextN(count: number) {
  if (!rotator) {
    console.log('⚠ Please initialize the team first using "init" command');
    return;
  }

  try {
    const members = rotator.nextN(count);
    console.log(`→ Next ${count} members:`);
    members.forEach((m, i) => {
      console.log(`  ${i + 1}. ${m.name} (id: ${m.id})`);
    });
  } catch (error) {
    if (error instanceof Error) {
      console.log(`⚠ Error: ${error.message}`);
    }
  }
}

// Reset rotator
function resetRotator() {
  if (!rotator) {
    console.log('⚠ Rotator not initialized');
    return;
  }

  rotator.reset();
  console.log('✓ Rotator reset to beginning');
}

// Process command
function processCommand(input: string) {
  const parts = input.trim().split(/\s+/);
  const command = parts[0].toLowerCase();
  const args = parts.slice(1);

  switch (command) {
    case 'init':
      if (args.length === 0) {
        console.log('⚠ Usage: init <members>');
        console.log('  Example: init Alice,Bob,Charlie');
        break;
      }
      initTeam(args[0].split(','));
      break;

    case 'init-with':
      if (args.length < 2) {
        console.log('⚠ Usage: init-with <members> <lastId>');
        console.log('  Example: init-with Alice,Bob,Charlie 1');
        break;
      }
      const lastId = parseInt(args[1], 10);
      if (isNaN(lastId)) {
        console.log('⚠ Last ID must be a number');
        break;
      }
      initTeam(args[0].split(','), lastId);
      break;

    case 'add':
      if (args.length === 0) {
        console.log('⚠ Usage: add <name> [active]');
        console.log('  Example: add Alice');
        console.log('  Example: add Bob false');
        break;
      }
      const isActive = args.length > 1 ? args[1].toLowerCase() === 'true' : true;
      addMember(args[0], isActive);
      break;

    case 'next':
      if (args.length === 0) {
        getNext();
      } else {
        const count = parseInt(args[0], 10);
        if (isNaN(count) || count <= 0) {
          console.log('⚠ Count must be a positive number');
          break;
        }
        getNextN(count);
      }
      break;

    case 'show':
      showTeam();
      break;

    case 'status':
      showStatus();
      break;

    case 'reset':
      resetRotator();
      break;

    case 'help':
      displayHelp();
      break;

    case 'exit':
    case 'quit':
      console.log('\nGoodbye! 👋\n');
      rl.close();
      process.exit(0);
      break;

    case '':
      // Empty command, just show prompt again
      break;

    default:
      console.log(`⚠ Unknown command: ${command}`);
      console.log('  Type "help" for available commands');
  }
}

// Main
displayWelcome();
rl.prompt();

rl.on('line', (input) => {
  processCommand(input);
  rl.prompt();
});

rl.on('close', () => {
  console.log('\nGoodbye! 👋\n');
  process.exit(0);
});