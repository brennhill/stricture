// src/api/types.ts
export interface ApiResponse {
  status: number;
  data: {
    userId: string;
    userName: string;
    userEmail: string;
    createdAt: string;
  };
  meta: {
    requestId: string;
    timestamp: number;
  };
}
