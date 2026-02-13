// test/data-transformer.anti4.test.ts (BAD)
import { describe, it, expect } from 'vitest';
import { DataTransformer } from '../src/data-transformer';

describe('DataTransformer', () => {
  it('should work', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([{ id: 1 }, { id: 2 }]);
    expect(result.records).toHaveLength(2);
  });

  it('test 1', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([]);
    expect(result.records).toHaveLength(0);
  });

  it('test 2', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([{ id: 1 }]);
    expect(result.stats.total).toBe(1);
  });

  it('handles data', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([null, { id: 1 }]);
    expect(result.errors).toHaveLength(1);
  });

  it('it works correctly', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([{ id: 1 }, null, { id: 2 }]);
    expect(result.records.length).toBe(2);
  });

  it('should handle edge cases', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([undefined]);
    expect(result.errors.length).toBeGreaterThan(0);
  });

  it('test transform function', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([{ name: 'test' }]);
    expect(result.stats.successful).toBe(1);
  });
});
