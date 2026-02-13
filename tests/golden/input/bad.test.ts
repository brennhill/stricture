// bad.test.ts — Test file with TQ violations.

import { getUser } from './bad';

describe('getUser', () => {
  it('works', async () => {
    const result = await getUser('123');
    expect(result).toBeDefined();  // TQ-no-shallow-assertions
    expect(result).toBeTruthy();   // TQ-no-shallow-assertions
  });
  // TQ-negative-cases: no error test
  // TQ-error-path-coverage: no error path tested
});
