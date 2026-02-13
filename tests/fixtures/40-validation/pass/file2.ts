// tests/utils/pagination.test.ts
import { PaginationUtil } from '../../src/utils/pagination';

describe('PaginationUtil', () => {
  // TQ-test-isolation: Create fresh data in each test
  const createTestData = () => {
    return Array.from({ length: 100 }, (_, i) => ({
      id: `item-${i + 1}`,
      name: `Item ${i + 1}`,
    }));
  };

  describe('validate', () => {
    // TQ-boundary-tested: Test all boundary values for page
    it('should accept page=1 (minimum valid)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 1, pageSize: 10 });
      }).not.toThrow();
    });

    it('should accept page=100 (maximum valid)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 100, pageSize: 10 });
      }).not.toThrow();
    });

    it('should reject page=0 (below minimum)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 0, pageSize: 10 });
      }).toThrow('Page must be between 1 and 100');
    });

    it('should reject page=101 (above maximum)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 101, pageSize: 10 });
      }).toThrow('Page must be between 1 and 100');
    });

    // TQ-boundary-tested: Test all boundary values for pageSize
    it('should accept pageSize=1 (minimum valid)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 1, pageSize: 1 });
      }).not.toThrow();
    });

    it('should accept pageSize=50 (maximum valid)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 1, pageSize: 50 });
      }).not.toThrow();
    });

    it('should reject pageSize=0 (below minimum)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 1, pageSize: 0 });
      }).toThrow('Page size must be between 1 and 50');
    });

    it('should reject pageSize=51 (above maximum)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 1, pageSize: 51 });
      }).toThrow('Page size must be between 1 and 50');
    });
  });

  describe('paginate', () => {
    it('should return first page of results', () => {
      const items = createTestData();

      const result = PaginationUtil.paginate(items, { page: 1, pageSize: 10 });

      // TQ-return-type-verified: All fields of PaginatedResponse verified
      expect(result.items).toHaveLength(10);
      expect(result.total).toBe(100);
      expect(result.page).toBe(1);
      expect(result.pageSize).toBe(10);
      expect(result.totalPages).toBe(10);

      // TQ-schema-conformance: Verify exact schema
      expect(Object.keys(result).sort()).toEqual(['items', 'page', 'pageSize', 'total', 'totalPages'].sort());

      // TQ-assertion-depth: Verify actual item content
      expect(result.items[0].id).toBe('item-1');
      expect(result.items[0].name).toBe('Item 1');
      expect(result.items[9].id).toBe('item-10');
      expect(result.items[9].name).toBe('Item 10');
    });

    it('should return middle page of results', () => {
      const items = createTestData();

      const result = PaginationUtil.paginate(items, { page: 5, pageSize: 10 });

      expect(result.items).toHaveLength(10);
      expect(result.page).toBe(5);
      expect(result.items[0].id).toBe('item-41');
      expect(result.items[9].id).toBe('item-50');
    });

    it('should return last page of results', () => {
      const items = createTestData();

      const result = PaginationUtil.paginate(items, { page: 10, pageSize: 10 });

      expect(result.items).toHaveLength(10);
      expect(result.page).toBe(10);
      expect(result.items[0].id).toBe('item-91');
      expect(result.items[9].id).toBe('item-100');
    });

    it('should handle partial last page', () => {
      const items = createTestData().slice(0, 25);

      const result = PaginationUtil.paginate(items, { page: 3, pageSize: 10 });

      expect(result.items).toHaveLength(5);
      expect(result.total).toBe(25);
      expect(result.totalPages).toBe(3);
      expect(result.items[0].id).toBe('item-21');
      expect(result.items[4].id).toBe('item-25');
    });

    it('should throw error when page is invalid', () => {
      const items = createTestData();

      expect(() => {
        PaginationUtil.paginate(items, { page: 0, pageSize: 10 });
      }).toThrow('Page must be between 1 and 100');
    });
  });
});
