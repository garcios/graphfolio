import { createClient as createGenqlClient, type Client } from './generated';

export interface ClientConfig {
  url?: string;
  headers?: Record<string, string> | (() => Promise<Record<string, string>> | Record<string, string>);
}

export const DEFAULT_BFF_URL = 'http://localhost:8080/query';

let currentAccessToken: string | null = null;

export function setClientAccessToken(token: string | null): void {
  currentAccessToken = token;
}

export function getClientAccessToken(): string | null {
  return currentAccessToken;
}

export function createGraphfolioClient(config: ClientConfig = {}): Client {
  return createGenqlClient({
    url: config.url || DEFAULT_BFF_URL,
    credentials: 'include',
    headers: async () => {
      const customHeaders =
        typeof config.headers === 'function'
          ? await config.headers()
          : (config.headers || {});
      const authHeaders: Record<string, string> = {};
      if (currentAccessToken) {
        authHeaders['Authorization'] = `Bearer ${currentAccessToken}`;
      }
      return {
        ...authHeaders,
        ...customHeaders,
      };
    },
  });
}

// Default singleton client instance for applications
export const client: Client = createGraphfolioClient();
