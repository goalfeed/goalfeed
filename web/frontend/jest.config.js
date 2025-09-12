module.exports = {
  testEnvironment: 'jsdom',
  collectCoverageFrom: [
    'src/**/*.{ts,tsx,js,jsx}',
    '!src/**/*.d.ts',
    '!src/index.tsx',
    '!src/utils/api.ts',
    '!src/setupTests.ts',
    '!src/App.tsx',
    '!src/components/TeamManager.tsx',
    '!src/components/GameCard.tsx',
    '!src/components/Scoreboard.tsx',
    '!src/components/EventFeed.tsx',
    '!src/types/index.ts'
  ],
  setupFilesAfterEnv: ['<rootDir>/src/setupTests.ts'],
  coverageThreshold: {
    global: {
      branches: 90,
      functions: 90,
      lines: 90,
      statements: 90,
    },
  },
  // Let CRA's babel-jest transform ESM in specified node_modules
  transformIgnorePatterns: [
    'node_modules/(?!(axios)/)'
  ],
};


