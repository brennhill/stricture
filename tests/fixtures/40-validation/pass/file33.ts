// src/storage/store.ts
export class Store {
  private data: Record<string, string> = {};

  set(key: string, value: string): void {
    this.data[key] = value;
  }

  get(key: string): string | undefined {
    return this.data[key];
  }
}
