// src/api/client.ts
export async function fetchData(url: string): Promise<Data> {
  try {
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    return response.json();
  } catch (error) {
    if (error instanceof TypeError) {
      throw new Error('Network error');
    }
    throw error;
  }
}
