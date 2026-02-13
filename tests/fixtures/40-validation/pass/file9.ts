// src/storage/cache.ts
export class Cache {
  private store = new Map<string, string>();

  set(key: string, value: string): { stored: boolean; key: string } {
    this.store.set(key, value);
    return { stored: true, key };
  }
}
