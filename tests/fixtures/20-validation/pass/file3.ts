// src/client/api-client.ts — Fetch-based API client for User API

import { User, CreateUserRequest, UpdateUserRequest, PaginatedResponse } from "../shared/types.js";

export class UserApiClient {
  constructor(private baseUrl: string, private authToken: string) {}

  async createUser(req: CreateUserRequest): Promise<User> {
    try {
      const response = await fetch(`${this.baseUrl}/api/users`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.authToken}`,
        },
        body: JSON.stringify(req),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || `HTTP ${response.status}`);
      }

      return await response.json() as User;
    } catch (err) {
      throw new Error(`Failed to create user: ${(err as Error).message}`);
    }
  }

  async getUser(id: string): Promise<User> {
    try {
      const response = await fetch(`${this.baseUrl}/api/users/${id}`, {
        headers: {
          Authorization: `Bearer ${this.authToken}`,
        },
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || `HTTP ${response.status}`);
      }

      return await response.json() as User;
    } catch (err) {
      throw new Error(`Failed to get user: ${(err as Error).message}`);
    }
  }

  async listUsers(cursor?: string, limit = 10): Promise<PaginatedResponse<User>> {
    try {
      const params = new URLSearchParams();
      if (cursor) params.set("cursor", cursor);
      params.set("limit", limit.toString());

      const response = await fetch(`${this.baseUrl}/api/users?${params}`, {
        headers: {
          Authorization: `Bearer ${this.authToken}`,
        },
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || `HTTP ${response.status}`);
      }

      return await response.json() as PaginatedResponse<User>;
    } catch (err) {
      throw new Error(`Failed to list users: ${(err as Error).message}`);
    }
  }

  async updateUser(id: string, req: UpdateUserRequest): Promise<User> {
    try {
      const response = await fetch(`${this.baseUrl}/api/users/${id}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.authToken}`,
        },
        body: JSON.stringify(req),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || `HTTP ${response.status}`);
      }

      return await response.json() as User;
    } catch (err) {
      throw new Error(`Failed to update user: ${(err as Error).message}`);
    }
  }
}
