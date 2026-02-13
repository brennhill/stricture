// test/pagination.anti8.test.ts (BAD)
import { describe, it, expect } from 'vitest';
import { paginate } from '../src/pagination';

describe('Pagination tests', () => {
  it('returns page 2 of results', () => {
    const data = Array.from({ length: 100 }, (_, i) => ({ id: i }));
    const result = paginate(data, { page: 2, limit: 10 });
    expect(result.items).toHaveLength(10);
    expect(result.page).toBe(2);
  });

  it('paginates with limit 20', () => {
    const data = Array.from({ length: 100 }, (_, i) => ({ id: i }));
    const result = paginate(data, { page: 3, limit: 20 });
    expect(result.items).toHaveLength(20);
  });

  it('handles page 5', () => {
    const data = Array.from({ length: 100 }, (_, i) => ({ id: i }));
    const result = paginate(data, { page: 5, limit: 15 });
    expect(result.items.length).toBeGreaterThan(0);
  });

  it('returns middle page', () => {
    const data = Array.from({ length: 200 }, (_, i) => ({ id: i }));
    const result = paginate(data, { page: 4, limit: 25 });
    expect(result.totalPages).toBe(8);
  });

  it('paginates typical dataset', () => {
    const data = Array.from({ length: 75 }, (_, i) => ({ id: i }));
    const result = paginate(data, { page: 2, limit: 10 });
    expect(result.hasNextPage).toBe(true);
  });

  it('works with standard limit', () => {
    const data = Array.from({ length: 100 }, (_, i) => ({ id: i }));
    const result = paginate(data, { page: 3, limit: 25 });
    expect(result.items).toHaveLength(25);
  });

  // Missing: page=0, page=1, page=-1, page=999999
  // Missing: limit=0, limit=1, limit=-1, limit=999999
  // Missing: empty data array
  // Missing: data.length < limit
  // Missing: page > totalPages
});
