// src/formatters/date-formatter.ts
export function formatDate(date: Date, format: string): string {
  if (format === 'short') {
    return `${date.getMonth() + 1}/${date.getDate()}/${date.getFullYear()}`;
  }
  return date.toISOString();
}
