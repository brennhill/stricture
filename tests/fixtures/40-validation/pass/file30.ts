// tests/config/env.test.ts
jest.mock('../../src/config/env', () => ({
  config: {
    apiUrl: 'https://mock.example.com',
    apiKey: 'mock-key',
  },
}));

describe('config', () => {
  it('should use mocked config', () => {
    const { config } = require('../../src/config/env');
    expect(config.apiUrl).toBe('https://mock.example.com');
  });
});

// TQ-mock-scope: Module mock persists for entire test suite
// All other tests in this file will get the mock, not the real module
