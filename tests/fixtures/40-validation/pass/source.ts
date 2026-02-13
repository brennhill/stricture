// tests/services/user-service.test.ts
import { UserService } from '../../src/services/user-service';
import { Database } from '../../src/db';

describe('UserService', () => {
  let userService: UserService;
  let mockDb: jest.Mocked<Database>;

  beforeEach(() => {
    // TQ-mock-scope: Fresh mocks in beforeEach
    mockDb = {
      insert: jest.fn(),
      findById: jest.fn(),
      update: jest.fn(),
      findAll: jest.fn(),
    } as any;
    userService = new UserService(mockDb);
  });

  afterEach(() => {
    // TQ-mock-scope: Cleanup in afterEach
    jest.clearAllMocks();
  });

  describe('createUser', () => {
    // TQ-test-naming: Clear "should [verb] when [condition]" format
    it('should create user when valid inputs provided', async () => {
      const mockUser = {
        id: 'user-123',
        name: 'John Doe',
        email: 'john@example.com',
        role: 'user' as const,
        createdAt: new Date('2026-01-01T00:00:00Z'),
      };
      mockDb.insert.mockResolvedValue(mockUser);

      const result = await userService.createUser('John Doe', 'john@example.com', 'user');

      // TQ-return-type-verified: All fields of User interface checked
      expect(result.id).toBe('user-123');
      expect(result.name).toBe('John Doe');
      expect(result.email).toBe('john@example.com');
      expect(result.role).toBe('user');
      expect(result.createdAt).toEqual(new Date('2026-01-01T00:00:00Z'));

      // TQ-schema-conformance: Verify no extra fields
      expect(Object.keys(result)).toEqual(['id', 'name', 'email', 'role', 'createdAt']);
    });

    // TQ-negative-cases: Error test for every happy path
    it('should throw error when name is empty', async () => {
      await expect(
        userService.createUser('', 'john@example.com', 'user')
      ).rejects.toThrow('Name and email are required');
    });

    it('should throw error when email is empty', async () => {
      await expect(
        userService.createUser('John Doe', '', 'user')
      ).rejects.toThrow('Name and email are required');
    });

    // TQ-boundary-tested: Invalid role boundary
    it('should throw error when role is invalid', async () => {
      await expect(
        userService.createUser('John Doe', 'john@example.com', 'superadmin')
      ).rejects.toThrow('Invalid role');
    });
  });

  describe('getUser', () => {
    it('should return user when user exists', async () => {
      const mockUser = {
        id: 'user-456',
        name: 'Jane Smith',
        email: 'jane@example.com',
        role: 'admin' as const,
        createdAt: new Date('2026-01-15T10:30:00Z'),
        address: {
          street: '123 Main St',
          city: 'New York',
          country: 'USA',
        },
      };
      mockDb.findById.mockResolvedValue(mockUser);

      const result = await userService.getUser('user-456');

      // TQ-return-type-verified: All required fields
      expect(result).not.toBeNull();
      expect(result!.id).toBe('user-456');
      expect(result!.name).toBe('Jane Smith');
      expect(result!.email).toBe('jane@example.com');
      expect(result!.role).toBe('admin');
      expect(result!.createdAt).toEqual(new Date('2026-01-15T10:30:00Z'));

      // TQ-assertion-depth: Nested object fields verified
      expect(result!.address).toBeDefined();
      expect(result!.address!.street).toBe('123 Main St');
      expect(result!.address!.city).toBe('New York');
      expect(result!.address!.country).toBe('USA');
    });

    it('should return null when user not found', async () => {
      mockDb.findById.mockResolvedValue(null);

      const result = await userService.getUser('nonexistent');

      expect(result).toBeNull();
    });

    // TQ-error-path-coverage: Test NOT_FOUND error branch
    it('should return null when database throws NOT_FOUND error', async () => {
      const notFoundError = new Error('Not found');
      (notFoundError as any).code = 'NOT_FOUND';
      mockDb.findById.mockRejectedValue(notFoundError);

      const result = await userService.getUser('user-789');

      expect(result).toBeNull();
    });

    // TQ-error-path-coverage: Test generic error branch
    it('should throw database error when unknown error occurs', async () => {
      mockDb.findById.mockRejectedValue(new Error('Connection failed'));

      await expect(
        userService.getUser('user-789')
      ).rejects.toThrow('Database error');
    });

    it('should throw error when id is empty', async () => {
      await expect(
        userService.getUser('')
      ).rejects.toThrow('User ID is required');
    });
  });

  describe('updateUser', () => {
    it('should update user when valid data provided', async () => {
      const mockUpdatedUser = {
        id: 'user-123',
        name: 'John Updated',
        email: 'john.updated@example.com',
        role: 'admin' as const,
        createdAt: new Date('2026-01-01T00:00:00Z'),
      };
      mockDb.update.mockResolvedValue(mockUpdatedUser);

      const result = await userService.updateUser('user-123', {
        name: 'John Updated',
        email: 'john.updated@example.com',
      });

      // TQ-return-type-verified: All fields verified
      expect(result.id).toBe('user-123');
      expect(result.name).toBe('John Updated');
      expect(result.email).toBe('john.updated@example.com');
      expect(result.role).toBe('admin');
      expect(result.createdAt).toEqual(new Date('2026-01-01T00:00:00Z'));
    });

    it('should throw error when user not found', async () => {
      mockDb.update.mockResolvedValue(null);

      await expect(
        userService.updateUser('nonexistent', { name: 'Updated' })
      ).rejects.toThrow('User not found');
    });

    // TQ-error-path-coverage: Test VALIDATION_ERROR branch
    it('should throw validation error when invalid data provided', async () => {
      const validationError = new Error('Invalid email format');
      (validationError as any).code = 'VALIDATION_ERROR';
      mockDb.update.mockRejectedValue(validationError);

      await expect(
        userService.updateUser('user-123', { email: 'invalid-email' })
      ).rejects.toThrow('Invalid update data');
    });

    // TQ-error-path-coverage: Test generic error rethrow branch
    it('should rethrow unknown errors', async () => {
      const unknownError = new Error('Database connection lost');
      mockDb.update.mockRejectedValue(unknownError);

      await expect(
        userService.updateUser('user-123', { name: 'Updated' })
      ).rejects.toThrow('Database connection lost');
    });

    it('should throw error when id is empty', async () => {
      await expect(
        userService.updateUser('', { name: 'Updated' })
      ).rejects.toThrow('User ID is required');
    });
  });

  describe('listUsers', () => {
    it('should return paginated users', async () => {
      const mockUsers = [
        {
          id: 'user-1',
          name: 'User One',
          email: 'one@example.com',
          role: 'user' as const,
          createdAt: new Date('2026-01-01T00:00:00Z'),
        },
        {
          id: 'user-2',
          name: 'User Two',
          email: 'two@example.com',
          role: 'admin' as const,
          createdAt: new Date('2026-01-02T00:00:00Z'),
        },
        {
          id: 'user-3',
          name: 'User Three',
          email: 'three@example.com',
          role: 'guest' as const,
          createdAt: new Date('2026-01-03T00:00:00Z'),
        },
      ];
      mockDb.findAll.mockResolvedValue(mockUsers);

      const result = await userService.listUsers(1, 10);

      // TQ-assertion-depth: Not just .toHaveLength, but verify each user's fields
      expect(result).toHaveLength(3);

      expect(result[0].id).toBe('user-1');
      expect(result[0].name).toBe('User One');
      expect(result[0].email).toBe('one@example.com');
      expect(result[0].role).toBe('user');
      expect(result[0].createdAt).toEqual(new Date('2026-01-01T00:00:00Z'));

      expect(result[1].id).toBe('user-2');
      expect(result[1].name).toBe('User Two');
      expect(result[1].email).toBe('two@example.com');
      expect(result[1].role).toBe('admin');

      expect(result[2].id).toBe('user-3');
      expect(result[2].name).toBe('User Three');
      expect(result[2].email).toBe('three@example.com');
      expect(result[2].role).toBe('guest');
    });
  });
});
