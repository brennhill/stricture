// src/services/counter-service.ts
export class CounterService {
  private count = 0;

  increment(): number {
    this.count++;
    return this.count;
  }

  getCount(): number {
    return this.count;
  }
}
