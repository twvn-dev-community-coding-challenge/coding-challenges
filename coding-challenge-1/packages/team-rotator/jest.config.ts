import type { Config } from '@jest/types';
import nxPreset from '@nx/jest/preset';

const config: Config.InitialOptions = {
  displayName: 'team-rotator',
  ...nxPreset,
  testEnvironment: 'node',
  transform: {
    '^.+\\.[tj]s$': ['ts-jest', { tsconfig: '<rootDir>/tsconfig.spec.json' }],
  },
  moduleFileExtensions: ['ts', 'js', 'html'],
  coverageDirectory: '../../coverage/packages/team-rotator',
  collectCoverageFrom: [
    'src/**/*.ts',
    '!src/**/*.spec.ts',
    '!src/**/*.test.ts',
    '!src/index.ts', // Just re-exports
  ],
  coverageThreshold: {
    global: {
      branches: 82, // Some defensive branches are difficult to test in practice
      functions: 90,
      lines: 90,
      statements: 90,
    },
  },
};

export default config;