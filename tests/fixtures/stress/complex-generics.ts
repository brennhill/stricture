// complex-generics.ts — Stress test: complex generic type parameters.

export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

export type UnionToIntersection<U> = (U extends any ? (k: U) => void : never) extends (k: infer I) => void ? I : never;

export type Flatten<T> = T extends Array<infer U> ? U : T;

export type DeepReadonly<T> = {
  readonly [P in keyof T]: T[P] extends object ? DeepReadonly<T[P]> : T[P];
};

export interface ApiResponse<T, E = Error> {
  data: T | null;
  error: E | null;
  metadata: {
    requestId: string;
    timestamp: Date;
    pagination?: {
      page: number;
      total: number;
      hasMore: boolean;
    };
  };
}

export function processResponse<T extends Record<string, unknown>, E extends Error>(
  response: ApiResponse<T, E>,
  transform: (data: T) => DeepPartial<T>,
  onError?: (error: E) => void,
): DeepReadonly<ApiResponse<DeepPartial<T>, E>> {
  if (response.error) {
    onError?.(response.error);
    return { data: null, error: response.error, metadata: response.metadata } as any;
  }
  const transformed = transform(response.data!);
  return { data: transformed, error: null, metadata: response.metadata } as any;
}
